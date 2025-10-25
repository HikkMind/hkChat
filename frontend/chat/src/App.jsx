import React, { useState, useEffect, useRef } from 'react';
import LoginPage from './pages/loginPage';
import RegisterPage from './pages/registerPage';
import ChatListPage from './pages/chatlistPage';
import ChatPage from './pages/chatPage';
import useWebSocket from './hooks/websocket';
import useScrollToBottom from './hooks/scrollbottom';
import routes from `./constant/routes`
import {verifyAccessToken} from './hooks/verifytoken';



function App() {
  const [page, setPage] = useState(routes.login);
  const [currentUser, setCurrentUser] = useState(null);
  const [messages, setMessages] = useState([]);
  const [chats, setChats] = useState([]);         
  const [selectedChat, setSelectedChat] = useState(null);
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
      setCurrentUser(null);
      setPage(routes.login);
    },
    routes
  });

  useEffect(() => {
    verifyAccessToken(setCurrentUser, setPage, routes);
  }, []);

  const login = async (username, password) => {
    try {
      const response = await fetch('/login', {
        method: 'POST',
        credentials: "include",
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password })
      });

      if (response.ok) {
        const data = await response.json();

        if (data.status === 'ok' && data.access_token) {
          setCurrentUser({ username: username, accessToken: data.access_token });
          setPage(routes.chatList);
          localStorage.setItem("accessToken", data.access_token);
          localStorage.setItem("username", username);
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

  const joinChat = (chat) => {
    setSelectedChat(chat);
    setMessages([]);
    setPage(routes.chat);
  }

  const leaveChat = () => {
    if (socketRef.current) {
      socketRef.current.send(JSON.stringify({
        intent: "leave_chat",
        chat_id: selectedChat.chat_id
      }))
    }
    setPage(routes.chatList)
  }

  const logout = () => {
    if (socketRef.current) socketRef.current.close();
    setCurrentUser(null);
    setMessages([]);
    setPage(routes.login);
    socketRef.current = null;
    localStorage.removeItem('accessToken');
    localStorage.removeItem('username');
  };

  const chatAction = (name, action) => {
    const text = name.trim()
    if (action == 'create_chat'){
      if (!text || text.length < 6) {
        return
      }
    }

    const msgObj = {
      intent: action,
      text: text
    };

    console.log(action, ' : ', name)
    
    sendMessage(msgObj);
  }


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
        onSelectChat={joinChat}
        onLogout={logout}
        onChatAction={chatAction}
        currentUser={currentUser}
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
        onChatlist={leaveChat}
      />
    );
  }

  return null;
}

export default App;
