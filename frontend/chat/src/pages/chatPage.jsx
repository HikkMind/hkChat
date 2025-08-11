import React from 'react';
import { useState } from 'react';

export default function ChatPage({ currentChat, messages, onSendMessage, messageInputRef, chatRef, onLogout }) {
  
  const handleSend = () => {
    const text = messageInputRef.current.value.trim();
    if (!text) return;

    const msgObj = {
      intent: 'send_message',
      text: text
    };
    
    onSendMessage(msgObj);
    
    messageInputRef.current.value = '';
  };
  
  const handleKeyPress = (e) => {
    if (e.key === 'Enter') {
      handleSend();
    }
  };
  return (
    <div>
      <h2>Чат (<span>{currentChat?.chat_name}</span>)</h2>
      <div
        id="chat"
        ref={chatRef}
        style={{ height: 300, border: '1px solid #ccc', overflowY: 'scroll', padding: 8, marginBottom: 8 }}
      >
        {messages.map((msg, i) => (
          <div key={i}>
            <strong>{msg.username}:</strong> <span>{msg.message}</span> <small>{new Date(msg.time).toLocaleTimeString()}</small>
          </div>
        ))}
      </div>
      <input
        id="message"
        placeholder="Сообщение"
        ref={messageInputRef}
        onKeyDown={handleKeyPress}
      />
      <button id="send-message-btn" onClick={handleSend}>Отправить</button>
      <button id="logout-btn" onClick={onLogout}>Выйти</button>
    </div>
  );
}