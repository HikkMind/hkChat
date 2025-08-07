import React from 'react';
import { useState } from 'react';

export default function ChatListPage({ chats, onSelectChat }) {
  return (
    <div>
      <h2>Список чатов</h2>
      <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
        {chats.map(chat => (
          <div
            key={chat.id}
            onClick={() => onSelectChat(chat)}
            style={{
              padding: '10px',
              border: '1px solid #ccc',
              borderRadius: '8px',
              cursor: 'pointer',
              backgroundColor: '#f9f9f9'
            }}
          >
            {chat.name}
          </div>
        ))}
      </div>
    </div>
  );
}