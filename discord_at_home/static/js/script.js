
async function userSignUp() {
    var formData = new FormData();
    var userName = document.getElementById("username").value;
    var password = document.getElementById("password").value;
    console.log(userName);
    formData.append('username', userName);
    formData.append('password', password);

    const response = await fetch("http://localhost:8080/register", {
        method: "POST",
        body: formData,
        headers: {
        'Origin': 'http://localhost:3000',
        },
    });

    if (response.status == 201) {
        console.log("Created user successfully")
    }
    else {
        console.log("Failed to create user")
        openModal();
    }

    if (response.ok) {
        const data = await response.text();
        console.log("Data sent to server");
        console.log(data);
        window.location.href = "signupsuccess.html";
        //return data;
    }
    else {
        console.log("Error sending data to the server");
        throw new Error(`Error sending data: ${response.statusText}`);
    }
}

function openModal() {
    document.getElementById('myModal').style.display = 'flex';
}

function closeModal() {
    document.getElementById('myModal').style.display = 'none';
}

async function Login() {
    var formData = new FormData();
    var username = document.getElementById("username").value;
    var password = document.getElementById("password").value;
    formData.append('username', username);
    formData.append('password', password);

    const response = await fetch("http://localhost:8080/login", {
        method: "POST",
        body: formData,
        headers: {
            'Origin': 'http://localhost:3000',
        }
    });

    if (response.status == 202) {
        document.cookie = 'username=' + username;
        console.log("Status 202");
    } else {
        console.log("Failed to login");
    }

    if (response.ok) {
        const data = await response.text();
        console.log(data);
        window.location.href = "chats.html";
    }
    else {
        console.log("Error sending data to the server");
        throw new Error(`Error sending data: ${response.statusText}`);
    }
}


