let socket;
let currentUser = null;

const routes = {
    '/login': 'login-page',
    '/register': 'register-page',
    '/chat': 'chat-page'
};

function showPage(pageId, push = true) {
    document.querySelectorAll('.page').forEach(page => {
        page.classList.remove('active');
    });

    document.getElementById(pageId).classList.add('active');

    if (push) {
        const path = Object.entries(routes).find(([_, v]) => v === pageId)?.[0];
        if (path) history.pushState({}, '', path);
    }
}
window.onpopstate = () => {
    
    const pageId = routes[location.pathname] || 'login-page';
    showPage(pageId, false);
};
window.addEventListener('DOMContentLoaded', () => {
    const pageId = routes[location.pathname];

    if (pageId) {
        showPage(pageId, false); 
    } else {
        showPage('login-page', false);
        history.replaceState({}, '', '/login');
    }
});
async function login() {
    const username = document.getElementById('login-username').value;
    const password = document.getElementById('login-password').value;
    
    const response = await fetch('/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password })
    });
    
    if (response.ok) {
        currentUser = username;
        document.getElementById('current-user').textContent = username;
        initWebSocket();
        showPage('chat-page');
    } else {
        alert('Ошибка авторизации');
    }
}

async function register() {
    const username = document.getElementById('reg-username').value;
    const password = document.getElementById('reg-password').value;
    
    const response = await fetch('/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password })
    });
    
    if (response.ok) {
        alert('Регистрация успешна!');
        showPage('login-page');
    } else {
        alert('Ошибка регистрации');
    }
}

function logout() {
    if (socket) socket.close();
    currentUser = null;
    showPage('login-page');
}

function initWebSocket() {
    socket = new WebSocket(`ws://localhost:8080/messager`);
    
    socket.onmessage = (event) => {
        const msg = JSON.parse(event.data);
        displayMessage(msg);
    };
    
    socket.onclose = () => {
        if (currentUser) {
            setTimeout(initWebSocket, 1000); // Переподключение
        }
    };
}

function sendMessage(event) {
    if (event instanceof KeyboardEvent && event.key !== 'Enter') {
        return;
    }
    if (!socket || !currentUser) return;
    
    const message = {
        username: currentUser,
        message: document.getElementById('message').value,
    };
    
    socket.send(JSON.stringify(message));
    document.getElementById('message').value = '';
}

function displayMessage(msg) {
    const time = new Date(msg.time).toLocaleTimeString();
    const messageElement = document.createElement('div');
    messageElement.innerHTML = `
        <strong>${msg.username}:</strong>
        <span>${msg.message}</span>
        <small>${time}</small>
    `;
    document.getElementById('chat').appendChild(messageElement);
    chat.scrollTop = chat.scrollHeight;
}

export { login, register, showPage, sendMessage, logout };
