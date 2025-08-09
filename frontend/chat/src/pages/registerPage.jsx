import React from 'react';
import { useState } from 'react';

export default function RegisterPage({ onRegister, onShowLogin }) {
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