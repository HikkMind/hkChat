let socket;
let currentUser = null;

// Переключение между страницами
function showPage(pageId) {
    document.querySelectorAll('.page').forEach(page => {
        page.classList.remove('active');
    });
    document.getElementById(pageId).classList.add('active');
}

// Авторизация
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

// Регистрация
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

// Выход
function logout() {
    if (socket) socket.close();
    currentUser = null;
    showPage('login-page');
}

// Инициализация WebSocket
function initWebSocket() {
    socket = new WebSocket(`ws://localhost:8080/ws?token=${currentUser}`);
    
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

// Отправка сообщения
function sendMessage() {
    if (!socket || !currentUser) return;
    
    const message = {
        sender: currentUser,
        message: document.getElementById('message').value,
        time: new Date().toISOString()
    };
    
    socket.send(JSON.stringify(message));
    document.getElementById('message').value = '';
}

// Отображение сообщения
function displayMessage(msg) {
    const time = new Date(msg.time).toLocaleTimeString();
    const messageElement = document.createElement('div');
    messageElement.innerHTML = `
        <strong>${msg.sender}:</strong>
        <span>${msg.message}</span>
        <small>${time}</small>
    `;
    document.getElementById('chat').appendChild(messageElement);
}