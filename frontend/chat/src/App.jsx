import React, { useState, useEffect, useRef } from 'react';

const routes = {
  login: 'login',
  register: 'register',
  chat: 'chat',
  chatList: 'chatlist'
};


function App() {
  const [page, setPage] = useState(routes.login);
  const [currentUser, setCurrentUser] = useState(null);
  const [messages, setMessages] = useState([]);
  const [chats, setChats] = useState([]);         
  const [selectedChat, setSelectedChat] = useState(null);
  const socketRef = useRef(null);
  const messageInputRef = useRef(null);

  useEffect(() => {
    if (!socketRef.current && page === routes.chatList) {
      const socket = new WebSocket(`ws://${window.location.hostname}:5173/chatlist`);
      socketRef.current = socket;

      socket.onopen = () => {
        if (currentUser) {
          socket.send(JSON.stringify({
            intent: "get_chats",
          }));
        }
      }

      socket.onclose = () => {
        if (currentUser) {
          setTimeout(() => {
            if (currentUser) setPage(routes.chat);
          }, 1000);
        }
      };

      socket.onmessage = (event) => {
        const msg = JSON.parse(event.data);
        if (msg.)
      };
    }
  }, [page, currentUser]);

  useEffect(() => {
    const socket = socketRef.current;
    if (page === routes.chat && currentUser && socket) {
      const joinMessage = JSON.stringify({
        intent: 'join_chat',
        chatId: selectedChat.id,
        name: selectedChat.name,
        userId: currentUser.id,
      });
      if (socket.readyState === WebSocket.OPEN) {
        socket.send(joinMessage);
      } 
      // else {
      //   socket.onopen = () => socket.send(joinMessage)
      // }

      // socket.onmessage = (event) => {
      //   const msg = JSON.parse(event.data);
      //   setMessages(prev => [...prev, msg]);
      // };
    }

  }, [page, currentUser, selectedChat]);
 
  // Закрываем сокет один раз — при размонтировании (выход из приложения)
  useEffect(() => {
    return () => {
      if (socketRef.current) {
        socketRef.current.close();
      }
    };
  }, []);


  const sendMessage = () => {
    if (!socketRef.current || !currentUser) return;

    const message = {
      intent: 'send_message',
      username: currentUser.username,
      message: messageInputRef.current.value
    };

    socketRef.current.send(JSON.stringify(message));
    messageInputRef.current.value = '';
  };

  const chatRef = useRef(null);
  useEffect(() => {
    if (chatRef.current) {
      chatRef.current.scrollTop = chatRef.current.scrollHeight;
    }
  }, [messages]);

  // const login = async (username, password) => {
  //   const response = await fetch('/login', {
  //     method: 'POST',
  //     headers: { 'Content-Type': 'application/json' },
  //     body: JSON.stringify({ username, password })
  //   });  
  const login = async (username, password) => {
    try {
      const response = await fetch('/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password })
      });

      if (response.ok) {
        const data = await response.json();

        if (data.status === 'ok' && data.userId) {
          setCurrentUser({ username: username, id: data.userId });
          setPage(routes.chatList);
        } else {
          alert('Неверный ответ от сервера');
        }
      } else {
        alert('Ошибка авторизации');
      }
    } catch (error) {
      console.error('Ошибка сети:', error);
      alert('Сервер недоступен');
    }
  };
  //   if (response.ok) {
  //     setCurrentUser({ name: username, id: data.userId });
  //     // setCurrentUser(username);
  //     // setPage(routes.chat);
  //     setPage('chatList')
  //   } else {
  //     alert('Ошибка авторизации');
  //   }
  // };

  const register = async (username, password) => {
    const response = await fetch('/register', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password })
    });
    if (response.ok) {
      alert('Регистрация успешна!');
      setPage(routes.login);
    } else {
      alert('Ошибка регистрации');
    }
  };

  const logout = () => {
    if (socketRef.current) socketRef.current.close();
    setCurrentUser(null);
    setMessages([]);
    setPage(routes.login);
  };


  if (page === routes.login) {
    return <LoginPage onLogin={login} onShowRegister={() => setPage(routes.register)} />;
  }

  if (page === routes.register) {
    return <RegisterPage onRegister={register} onShowLogin={() => setPage(routes.login)} />;
  }

  if (page === routes.chatList) {

    return (
      <ChatListPage
        chats={chats}
        onSelectChat={(chat) => {
          setSelectedChat(chat);
          setMessages([]);         // очищаем сообщения
          setPage(routes.chat);    // переходим на страницу чата
        }}
        onLogout={logout}
      />
    );
  }


  if (page === routes.chat) {
    return (
      <ChatPage
        currentChat={selectedChat}
        messages={messages}
        onSendMessage={sendMessage}
        messageInputRef={messageInputRef}
        chatRef={chatRef}
        onLogout={logout}
      />
    );
  }

  return null;
}


function LoginPage({ onLogin, onShowRegister }) {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');

  return (
    <div>
      <h2>Вход</h2>
      <input placeholder="Логин" value={username} onChange={e => setUsername(e.target.value)} />
      <input type="password" placeholder="Пароль" value={password} onChange={e => setPassword(e.target.value)} />
      <button onClick={() => onLogin(username, password)}>Войти</button>
      <button onClick={onShowRegister}>Регистрация</button>
    </div>
  );
}

function RegisterPage({ onRegister, onShowLogin }) {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');

  return (
    <div>
      <h2>Регистрация</h2>
      <input placeholder="Логин" value={username} onChange={e => setUsername(e.target.value)} />
      <input type="password" placeholder="Пароль" value={password} onChange={e => setPassword(e.target.value)} />
      <button onClick={() => onRegister(username, password)}>Зарегистрироваться</button>
      <button onClick={onShowLogin}>Назад</button>
    </div>
  );
}

function ChatListPage({ chats, onSelectChat }) {
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


function ChatPage({ currentChat, messages, onSendMessage, messageInputRef, chatRef, onLogout }) {
  const handleKeyPress = (e) => {
    if (e.key === 'Enter') {
      onSendMessage();
    }
  };

  return (
    <div>
      <h2>Чат (<span>{currentChat?.name}</span>)</h2>
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
        onKeyPress={handleKeyPress}
      />
      <button id="send-message-btn" onClick={onSendMessage}>Отправить</button>
      <button id="logout-btn" onClick={onLogout}>Выйти</button>
    </div>
  );
}

export default App;
