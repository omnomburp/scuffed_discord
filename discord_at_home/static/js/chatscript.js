const chatHistoryElement = document.querySelector('.chat-history');
const chatInputElement = document.querySelector('.chat-input');
const CookieUserName = document.cookie.split('=')[1];
let changeServerListeners = [];

async function LoadServers() {
    console.log(CookieUserName);
    const response = await fetch(`http://localhost:8080/loadservers?id=${CookieUserName}`, {
        method: "GET",
        headers: {
            'Origin': 'http://localhost:3000',
        },
    });

    if (response.status == 202) {
        console.log("Server data fetched")
    }
    else {
        console.log("Failed to get servers");
        return;
    }

    if (response.ok) {
        const data = await response.json();
        for (let i = 0; i < data.length; ++i) {
            addServer(data[i]);
        }
    } else {
        throw new Error("Failed to load servers");
    }


}

async function CreateServer(name) {
    formData = new FormData();
    var serverName = name;
    formData.append('servername', serverName);
    formData.append('username', CookieUserName);

    const response = await fetch("http://localhost:8080/createserver", {
        method: "POST",
        body: formData,
        headers: {
            'Origin': 'http://localhost:3000',
        },
    });

    if (response.status === 201) {
        addServer(serverName);
    } else {
        console.log("Unable to create server");
        return;
    }
}

async function JoinServer(name) {
    formData = new FormData();
    var serverName = name;
    formData.append('servername', serverName);
    formData.append('username', CookieUserName);

    const response = await fetch("http://localhost:8080/joinserver", {
        method: "POST",
        body: formData,
        headers: {
            'Origin': 'http://localhost:3000',
        },
    });

    if (response.status === 201) {
        addServer(serverName);
    } else {
        console.log("Unable to join server");
        return;
    }
}

function addServer(serverName) {
    const serverContainer = document.querySelector('.server-list');
    const serverLi = document.createElement('li');
    serverLi.classList.add('server');
    const iconSpan = document.createElement('span');
    iconSpan.classList.add('server-icon');
    iconSpan.innerHTML = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="24" height="24"><path fill="none" d="M0 0h24v24H0z"/><path d="M20 6h-8l-2-2H6c-1.1 0-1.99.89-1.99 2L4 20c0 1.1.89 2 2 2h16c1.1 0 2-.9 2-2V8l-2-2zm0 12H4V8l2 2h8l2-2v10zm-8-2h2v-4h-2v4zm-4 0h2v-4h-2v4z"/></svg>`
    const serverSpan = document.createElement('span');
    serverSpan.classList.add('server-name');
    serverSpan.textContent = serverName;
    changeServerListeners.push(serverSpan);
    
    serverLi.appendChild(iconSpan);
    serverLi.appendChild(serverSpan);
    serverLi.addEventListener('click', changeServer(serverSpan));
    serverContainer.appendChild(serverLi);

    StartNewEventSource(serverName);
}

async function LoadChat(servername) {
    const response = await fetch(`http://localhost:8080/loadchat?id=${servername}`, {
        method: "GET",
        headers: {
            'Origin': 'http://localhost:3000',
        }
    });

    if (response.status === 202) {
        console.log("response 202");
        const data = await response.json();
        data.forEach((message, index) => {
            addChatMessage(message.content, message.user);
        });
    } else {
        console.log("Failed to receive chat data");
        return;
    }    
}

async function sendMessage() {
    var formData = new FormData();
    var message = chatInputElement.value;
    formData.append('servername', currentServer)
    formData.append('content', message);
    formData.append('username', CookieUserName);

    const response = await fetch("http://localhost:8080/sendmessage", {
        method: "POST",
        body: formData,
        headers: {
            'Origin': 'http://localhost:3000'
        },
    });

    if (response.status === 202) {
        console.log("Message sent");
    } else {
        console.log("Message failed to send");
    }
}

function addChatMessage(message, username) {
    const messageElement = document.createElement('div');
    messageElement.classList.add('chat-message');
  
    const usernameElement = document.createElement('span');
    usernameElement.classList.add('chat-username');
    usernameElement.textContent = username + "  ";
  
    const messageContentElement = document.createElement('span');
    messageContentElement.classList.add('chat-content');
    messageContentElement.textContent = message;
  
    messageElement.appendChild(usernameElement);
    messageElement.appendChild(messageContentElement);
  
    chatHistoryElement.appendChild(messageElement);
  
    chatHistoryElement.scrollTop = chatHistoryElement.scrollHeight;
  }
  
function justifyChatText() {
    const chatMessages = chatHistoryElement.querySelectorAll('.chat-message');

    const availableHeight = chatHistoryElement.offsetHeight;

    let totalMessageHeight = 0;
    for (const message of chatMessages) {
    totalMessageHeight += message.offsetHeight;
    }

    const remainingSpace = availableHeight - totalMessageHeight;

    for (const message of chatMessages) {
    message.style.paddingBottom = `${remainingSpace / chatMessages.length}px`;
    }
}
  
  
window.addEventListener('resize', justifyChatText);
chatHistoryElement.addEventListener('DOMNodeInserted', justifyChatText);

const chatForm = document.querySelector('.chat-form');
chatForm.addEventListener('submit', (event) => {
event.preventDefault();

sendMessage();

chatInputElement.value = '';
});

var currentServer = "";

let eventSources = [];

function StartNewEventSource(serverName) {
    const eventSource = new EventSource(`http://localhost:8080/sse?id=${serverName}`);

    eventSource.onmessage = function (event) {
        const message = JSON.parse(event.data);
        addChatMessage(message.content, message.user);
        console.log("Received message:", message);
    };

    eventSource.onerror = function (error) {
        console.error("EventSource failed:", error);
    };

    eventSources.push(eventSource);
}

function changeServer(element) {
    
    currentServer = element.textContent;
    
    while (chatHistoryElement.firstChild) {
        chatHistoryElement.removeChild(chatHistoryElement.firstChild);
    }
    LoadChat(element.textContent);
}

const createButton = document.getElementById('create-button');

createButton.addEventListener('click', () => {
const text = prompt('Enter Server name:');
CreateServer(text);
});

const joinButton = document.getElementById('join-button');

joinButton.addEventListener('click', () => {
const text = prompt('Enter Server name:');
JoinServer(text);
});