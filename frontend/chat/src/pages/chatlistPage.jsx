import React from 'react';
import { useState } from 'react';

export default function ChatListPage({ chats, onSelectChat, onLogout, onCreateChat }) {

  const [chatName, setChatName] = useState('');

  const handleCreate = () => {
    if (chatName.trim().length < 6) {
      alert('Название чата должно содержать минимум 6 символов');
      return;
    }
    onCreateChat(chatName.trim());
    setChatName('');
  };

  const handleKeyPress = (e) => {
    if (e.key === 'Enter') {
      handleCreate();
    }
  };

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
      <div style={{ marginTop: '10px', display: 'flex', gap: '10px' }}>
        <input
          type="text"
          placeholder="Введите название чата"
          value={chatName}
          onChange={(e) => setChatName(e.target.value)}
          onKeyDown={handleKeyPress}
          style={{ flexGrow: 0, padding: '6px', borderRadius: '6px', border: '1px solid #ccc', width: '250px' }}
        />
        <button id="create-btn" onClick={handleCreate}>Создать</button>
      </div>
      <button id="logout-btn" onClick={onLogout}>Выйти</button>
    </div>
  );
}