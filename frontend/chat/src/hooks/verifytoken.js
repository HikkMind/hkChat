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
        setCurrentUser({username: username, accessToken: token})
        setPage(routes.chatList);
        return true;
      }
      if (data.status === "refresh") {
        localStorage.setItem("accessToken", data.access_token);
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