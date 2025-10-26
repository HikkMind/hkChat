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

  function createWebsocketConnection() {
    if (!socketRef.current){
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const socket = new WebSocket(`${protocol}//${window.location.host}/chatlist`);

      socketRef.current = socket;

      socket.onopen = () => {
        if (currentUser) {
            socket.send(JSON.stringify({intent: 'auth', token: 'Bearer '+localStorage.getItem("accessToken")}))
        }
      };
    }
  }

  useEffect(() => {
    if (page === routes.chatList) {
      
      createWebsocketConnection()

      let socket = socketRef.current
      sendWhenOpen(socketRef.current, JSON.stringify({ intent: "get_chats" }));

      socket.onmessage = (event) => {
        
        const msg = JSON.parse(event.data);

        if (msg.intent === 'chat_list') {
          onChatsReceived(msg.chat_list);
        } else if (msg.intent === 'send_message') {
          onMessageReceived(msg);
        } else if (msg.intent === 'create_chat') {
          // console.log('creating chat: ', msg)
          onChatsReceived(prev => [...prev, msg.chat_info]);
        }else if (msg.intent === 'delete_chat') {
          // console.log('deleting chat: ', msg);
          onChatsReceived(prev => prev.filter(c => c.chat_id !== msg.chat_info.chat_id));
          // console.log('deleted chat: ', msg);
        }
      };

      socket.onclose = (event) => {
        socketRef.current = null
        if (event.code == 1006) {
          createWebsocketConnection()
          console.log('Webscoket reopen')
          return
        }
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
        // console.log("send join to chat : ", selectedChat)
        sendJoin();
      } else {
        // console.log("add event for chat : ", selectedChat)
        socketRef.current.addEventListener('open', sendJoin, { once: true });
      }


    }
  }, [page, selectedChat, currentUser, socketRef.current]);

  useEffect(() => {
    return () => {
      console.log("close socket")
      if (socketRef.current) {
        socketRef.current.close();
      }
    };
  }, []);

  const sendMessage = (msgObj) => {
    if (socketRef.current && socketRef.current.readyState === WebSocket.OPEN) {
      const msg = JSON.stringify(msgObj)
      // console.log("message: ", msg)
      socketRef.current.send(msg);
    }
  };

  return {
    sendMessage,
    socketRef,
  };
}
