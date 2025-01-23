

async function register(event) {
    event.preventDefault();
    const username = document.getElementById('register-username').value;
    const email = document.getElementById('register-email').value;
    const password = document.getElementById('register-password').value;

    try {
        const response = await fetch('/api/register', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, email, password })
        });
        

        if (response.ok) {
            const verifyBlock = document.createElement('div');
            verifyBlock.id = 'verifyBlock'; // Assign the ID to the div

            // Append it to the body or any other container element
            document.body.appendChild(verifyBlock);
            verifyBlock.innerHTML = `
            <h2>Please, verify your email:
            <input id="code" type="text" name="code" placeholder="Enter your verifying code">
            <button type="Submit" onClick="SendVerificationEmail()">Send verification code</button>`
            alert('Registration successful.');
            
        } else {
            alert('Registration failed. Please try again.');
        }

    } catch (error) {
        console.error('Error:', error);
    }
}

async function SendVerificationEmail(event) {
    // Prevent form submission if this function is called from a form button
    if (event) event.preventDefault();
    
    const email = document.getElementById('register-email').value;
    let code = document.getElementById('code').value;

    // Validate that both email and code are filled in
    if (!email || !code) {
        alert('Please enter both your email and verification code.');
        return;
    }

    code = Number(code);
    console.log(email, code)
    try {
        const response = await fetch('/api/verify', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, code })
        });

        if (response.ok) {
            const result = await response.json();
            alert('Verification code sent successfully.');
        
            // Handle additional steps if needed (e.g., update the UI)
        } else {
            const error = await response.json();
            alert('Failed to send verification code: ' + error.message);
        }
    } catch (error) {
        console.error('Error:', error);
        alert('An error occurred while sending the verification code.');
    }
}

// document.addEventListener('DOMContentLoaded', sendVerificationEmail);

async function login(event) {
    event.preventDefault();
    const email = document.getElementById('login-email').value;
    const password = document.getElementById('login-password').value;

    try {
        const response = await fetch('http://127.0.0.1:8080/api/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password })
        });

        if (response.ok) {
            const data = await response.json();

            if (data.is_verified === false) {
                alert('Your email is not verified.');
                return;
            }

            localStorage.setItem('token', data.token);
            localStorage.setItem('currentUser', JSON.stringify({ id: data.user_id, email }));
            window.location.href = '/index'; // Redirect to the main page
        } else {
            alert('Login failed. Please try again.');
        }
    } catch (error) {
        console.error('Error:', error);
    }
}


document.getElementById('register-form').addEventListener('submit', register);
document.getElementById('login-form').addEventListener('submit', login);