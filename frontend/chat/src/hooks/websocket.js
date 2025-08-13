import { useEffect, useRef } from 'react';

export default function useWebSocket({ page, currentUser, selectedChat, onChatsReceived, onMessageReceived, onUnauthorized, routes }) {
  const socketRef = useRef(null);

  function sendWhenOpen(socket, data) {
    if (socket.readyState === WebSocket.OPEN) {
      socket.send(data);
    } else {
      setTimeout(() => sendWhenOpen(socket, data), 50); // проверяем снова через 50 мс
    }
  }

  useEffect(() => {
    if (page === routes.chatList) {
      if (!socketRef.current){
        // const socket = new WebSocket(`ws://${window.location.hostname}:5173/chatlist`);
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const socket = new WebSocket(`${protocol}//${window.location.host}/chatlist`);

        socketRef.current = socket;

        socket.onopen = () => {
          if (currentUser) {
              socket.send(JSON.stringify({intent: 'auth', token: localStorage.getItem("accessToken")}))
          }
        };
      }

      let socket = socketRef.current
      sendWhenOpen(socketRef.current, JSON.stringify({ intent: "get_chats" }));

      socket.onmessage = (event) => {
        
        const msg = JSON.parse(event.data);
        console.log("new socket message : ", msg)

        if (msg.intent === 'chat_list') {
          onChatsReceived(msg.chat_list);
        } else if (msg.intent === 'send_message') {
          onMessageReceived(msg);
        }
      };

      socket.onclose = () => {
        console.warn('WebSocket закрыт');
      };
    }
  }, [page, currentUser]);

  useEffect(() => {
    if (page === routes.chat && currentUser && socketRef.current && selectedChat) {
      const joinMessage = JSON.stringify({
        intent: 'join_chat',
        chat_id: selectedChat.chat_id,
        name: selectedChat.name,
        userId: currentUser.id,
      });

      const sendJoin = () => socketRef.current.send(joinMessage);

      if (socketRef.current.readyState === WebSocket.OPEN) {
        sendJoin();
      } else {
        socketRef.current.addEventListener('open', sendJoin, { once: true });
      }


    }
  }, [page, selectedChat, currentUser]);

  useEffect(() => {
    return () => {
      if (socketRef.current) {
        socketRef.current.close();
      }
    };
  }, []);

  const sendMessage = (msgObj) => {
    if (socketRef.current && socketRef.current.readyState === WebSocket.OPEN) {
      socketRef.current.send(JSON.stringify(msgObj));
    }
  };

  return {
    sendMessage,
    socketRef,
  };
}
