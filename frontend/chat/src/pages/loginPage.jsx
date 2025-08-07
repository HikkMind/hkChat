import React from 'react';
import { useState } from 'react';

export default function LoginPage({ onLogin, onShowRegister }) {
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