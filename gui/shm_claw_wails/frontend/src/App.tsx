import {useState} from 'react';
import './App.css';
import {AskModel, QueryContext, StoreContext} from "../wailsjs/go/main/App";

type Message = {
    role: 'user' | 'assistant';
    content: string;
};

function App() {
    const [input, setInput] = useState('');
    const [messages, setMessages] = useState<Message[]>([]);
    const [loading, setLoading] = useState(false);
    
    // Fixed session ID for demo purposes
    const sessionID = "session_react_demo_01";

    const handleSend = async () => {
        if (!input.trim()) return;

        const userMsg = input.trim();
        setInput('');
        setLoading(true);

        // 1. Show user message instantly
        setMessages(prev => [...prev, { role: 'user', content: userMsg }]);

        try {
            // 2. Store user message in SHM
            await StoreContext(sessionID, "User: " + userMsg);

            // 3. Retrieve context from SHM (optional, here we just show how it connects)
            const contextPrompt = await QueryContext(sessionID);

            // 4. Send query to LLM via backend
            // In a real app, we might send the full contextPrompt. 
            // Here we'll just send the direct prompt for simplicity, but prepend system instructions
            const finalPrompt = `System: Answer accurately.\nContext:\n${contextPrompt}\nUser:\n${userMsg}`;
            const reply = await AskModel(finalPrompt);

            // 5. Store assistant message in SHM
            await StoreContext(sessionID, "Assistant: " + reply);

            // 6. Update UI
            setMessages(prev => [...prev, { role: 'assistant', content: reply }]);
        } catch (e) {
            setMessages(prev => [...prev, { role: 'assistant', content: `Error: ${e}` }]);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div id="App">
            <header className="app-header">
                <h1>shmClaw Terminal</h1>
                <p className="subtitle">High-Performance LLM Context Manager</p>
            </header>

            <main className="chat-container">
                <div className="message-list">
                    {messages.length === 0 && (
                        <div className="empty-state">No messages yet. Say hello!</div>
                    )}
                    {messages.map((msg, idx) => (
                        <div key={idx} className={`message-bubble ${msg.role}`}>
                            <strong>{msg.role === 'user' ? 'You' : 'shmClaw'}</strong>
                            <div className="message-content">{msg.content}</div>
                        </div>
                    ))}
                    {loading && (
                        <div className="message-bubble assistant loading">
                            <div className="message-content">Thinking...</div>
                        </div>
                    )}
                </div>

                <div className="input-area">
                    <input
                        type="text"
                        value={input}
                        onChange={(e) => setInput(e.target.value)}
                        onKeyDown={(e) => e.key === 'Enter' && handleSend()}
                        placeholder="Type a message to store in shared memory..."
                        disabled={loading}
                    />
                    <button onClick={handleSend} disabled={loading || !input.trim()}>
                        Send
                    </button>
                </div>
            </main>
        </div>
    )
}

export default App
