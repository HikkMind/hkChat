import React, { useState, useEffect, useRef } from 'react';

const routes = {
  login: 'login',
  register: 'register',
  chat: 'chat'
};


function App() {
  const [page, setPage] = useState(routes.login);
  const [currentUser, setCurrentUser] = useState(null);
  const [messages, setMessages] = useState([]);
  const socketRef = useRef(null);
  const messageInputRef = useRef(null);

  useEffect(() => {
    if (page === routes.chat && currentUser) {
      const socket = new WebSocket(`ws://localhost:8080/messager`);
      socketRef.current = socket;

      socket.onmessage = (event) => {
        const msg = JSON.parse(event.data);
        setMessages(prev => [...prev, msg]);
      };

      socket.onclose = () => {
        if (currentUser) {
          setTimeout(() => {
            if (currentUser) setPage(routes.chat);
          }, 1000);
        }
      };

      return () => socket.close();
    }
  }, [page, currentUser]);

  const sendMessage = () => {
    if (!socketRef.current || !currentUser) return;

    const message = {
      username: currentUser,
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

  const login = async (username, password) => {
    const response = await fetch('/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password })
    });
    if (response.ok) {
      setCurrentUser(username);
      setPage(routes.chat);
    } else {
      alert('Ошибка авторизации');
    }
  };

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

  if (page === routes.chat) {
    return (
      <ChatPage
        currentUser={currentUser}
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

function ChatPage({ currentUser, messages, onSendMessage, messageInputRef, chatRef, onLogout }) {
  const handleKeyPress = (e) => {
    if (e.key === 'Enter') {
      onSendMessage();
    }
  };

  return (
    <div>
      <h2>Чат (<span>{currentUser}</span>)</h2>
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
