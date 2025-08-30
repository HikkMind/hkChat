import React from 'react';
import { useState } from 'react';

export default function ChatListPage({ chats, onSelectChat, onLogout }) {
  return (
    <div>
      <h2>Список чатов</h2>
      <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
        { chats.length > 0 ? (
          chats.map(chat => (
            <div
              key={chat.chat_id}
              onClick={() => onSelectChat(chat)}
              style={{
                padding: '10px',
                border: '1px solid #ccc',
                borderRadius: '8px',
                cursor: 'pointer',
                backgroundColor: '#f9f9f9'
              }}
            >
              {chat.chat_name}
            </div>
          ))) : (
            <p style={{color: '#888' }}>
              Пока здесь нет чатов
            </p>
          )
        }
      </div>
      <button id="logout-btn" onClick={onLogout}>Выйти</button>
    </div>
  );
}