import { useEffect, useRef } from 'react';

export const verifyAccessToken = async (setCurrentUser, setPage, routes) => {
  try {
    const token = localStorage.getItem("accessToken");
    const username = localStorage.getItem("username");

    if (!token || !username) {
        return false;
    }

    const response = await fetch('/verifytoken', {
      method: 'POST',
      credentials: "include",
      headers: {
        "Authorization": `Bearer ${token}`
      }
    });

    if (response.ok) {
      const data = await response.json();
      if (data.status === "ok") {
        setCurrentUser({username: username, accessToken: data.accessToken})
        setPage(routes.chatList);
        return true;
      }
      if (data.status === "refresh") {
        localStorage.setItem("accessToken", data.accessToken);
        setCurrentUser({username: username, accessToken: data.accessToken})
        setPage(routes.chatList);
        return true;
      }
    }
  } catch (error) {
    console.error("verify token error : ", error);
  }
  return false;
};

// export const verifyToken = async (setCurrentUser, setPage, routes) => {
//   const token = localStorage.getItem("accessToken");
//   const username = localStorage.getItem("username");

//   if (!token || !username) return false;

//   try {
//     const res = await fetch('/verify', {
//       headers: { 'Authorization': `Bearer ${token}` }
//     });

//     if (res.ok) {
//       setCurrentUser({ username, accessToken: token });
//       setPage(routes.chatList);
//       return true;
//     } else if (res.status === 401) {
//       // пытаемся обновить
//       const refreshed = await refreshAccessToken();
//       if (refreshed) {
//         setCurrentUser({ username, accessToken: localStorage.getItem("accessToken") });
//         setPage(routes.chatList);
//         return true;
//       }
//     }
//   } catch (e) {
//     console.error("Ошибка проверки токена:", e);
//   }

//   return false;
// };