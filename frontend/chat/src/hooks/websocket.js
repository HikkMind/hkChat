import { useEffect, useRef } from 'react';

export default function useWebSocket({ page, currentUser, selectedChat, onChatsReceived, onMessageReceived, routes }) {
  const socketRef = useRef(null);

  useEffect(() => {
    if (!socketRef.current && page === routes.chatList) {
      const socket = new WebSocket(`ws://${window.location.hostname}:5173/chatlist`);
      socketRef.current = socket;

      socket.onopen = () => {
        if (currentUser) {
          socket.send(JSON.stringify({ intent: 'get_chats' }));
        }
      };

      socket.onmessage = (event) => {
        const msg = JSON.parse(event.data);

        if (msg.intent === 'chat_list') {
          onChatsReceived(msg.chats);
        } else if (msg.intent === 'new_message') {
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
        chatId: selectedChat.id,
        name: selectedChat.name,
        userId: currentUser.id,
      });

      if (socketRef.current.readyState === WebSocket.OPEN) {
        socketRef.current.send(joinMessage);
      } else {
        socketRef.current.onopen = () => {
          socketRef.current.send(joinMessage);
        };
      }
    }
  }, [page, selectedChat]);

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
