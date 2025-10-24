import React from 'react';
import { useState } from 'react';

export default function ChatListPage({ chats, onSelectChat, onLogout, onChatAction, currentUser }) {

  const [chatName, setChatName] = useState('');

  const handleCreate = () => {
    if (chatName.trim().length < 6) {
      alert('Название чата должно содержать минимум 6 символов');
      return;
    }
    onChatAction(chatName.trim(), 'create_chat');
    setChatName('');
  };

  const handleDelete = (chat_id) => {
    onChatAction(chat_id, 'delete_chat')
  }

  const handleKeyPress = (e) => {
    if (e.key === 'Enter') {
      handleCreate();
    }
  };

  console.log(...chats)

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
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                padding: '10px',
                border: '1px solid #ccc',
                borderRadius: '8px',
                cursor: 'pointer',
                backgroundColor: '#f9f9f9',
                position: 'relative',
              }}
            >
              <div 
                style={{
                display: 'flex',
                flexDirection: 'row',
                alignItems: 'center',
                flexGrow: 1,
                overflow: 'hidden',
                textOverflow: 'ellipsis',
                whiteSpace: 'nowrap',
                }}> 
                <span>{chat.chat_name}</span>
                {chat.owner_name && (
                  <span style={{ color: '#007bff', marginLeft: '5px' }}>
                    ({chat.owner_name})
                  </span>
                )} 
              </div>

              <button
              id='delete-btn'
              onClick={(e) => {
                e.stopPropagation();
                if (chat.owner_name === currentUser.username) {
                  handleDelete(String(chat.chat_id));
                }
              }}
              style={{
                flexShrink: 0,
                background: 'transparent',
                border: 'none',
                color: '#ff4d4f',
                fontSize: '18px',
                cursor: 'pointer',
                lineHeight: '1',
                marginLeft: '10px',
              }}
              title="Удалить чат"
            >
              ✕
            </button>
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