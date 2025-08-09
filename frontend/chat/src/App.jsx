import React, { useState, useEffect, useRef } from 'react';
import LoginPage from './pages/loginPage';
import RegisterPage from './pages/registerPage';
import ChatListPage from './pages/chatlistPage';
import ChatPage from './pages/chatPage';
import useWebSocket from './hooks/websocket';
import useScrollToBottom from './hooks/scrollbottom';
import routes from `./constant/routes`



function App() {
  const [page, setPage] = useState(routes.login);
  const [currentUser, setCurrentUser] = useState(null);
  const [messages, setMessages] = useState([]);
  const [chats, setChats] = useState([]);         
  const [selectedChat, setSelectedChat] = useState(null);
  // const socketRef = useRef(null);
  const messageInputRef = useRef(null);
  const chatRef = useScrollToBottom([messages]);

  const {
    sendMessage,
    socketRef
  } = useWebSocket({
    page,
    currentUser,
    selectedChat,
    onChatsReceived: setChats,
    onMessageReceived: (msg) => {
      setMessages(prev => [...prev, msg]);
    },
    onUnauthorized: () => {
      setCurrentUser(null);       // сбрасываем пользователя
      setPage(routes.login);      // переходим на страницу логина
    },
    routes
  });

  // const sendMessage = () => {
  //   if (!socketRef.current || !currentUser) return;

  //   const message = {
  //     intent: 'send_message',
  //     username: currentUser.username,
  //     message: messageInputRef.current.value
  //   };

  //   socketRef.current.send(JSON.stringify(message));
  //   messageInputRef.current.value = '';
  // };

  const login = async (username, password) => {
    try {
      const response = await fetch('/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password })
      });

      if (response.ok) {
        const data = await response.json();

        if (data.status === 'ok' && data.token) {
          setCurrentUser({ username: username, token: data.token });
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

export default App;
