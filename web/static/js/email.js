document.getElementById('emailForm').addEventListener('submit', async (e) => {
    e.preventDefault();

    const subject = document.getElementById('subject').value;
    const body = document.getElementById('body').value;
    const users = document.getElementById('users').value.split(',');

    try {
      const response = await fetch('http://127.0.0.1:8080/api/admin/broadcast-to-selected', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer YOUR_JWT_TOKEN', // Replace with the actual token
        },
        body: JSON.stringify({ subject, body, users }),
      });

      const result = await response.json(); // Try to parse the JSON response

      if (response.ok) {
        alert('Emails sent successfully!');
      } else {
        // Handle the error response
        alert(`Failed to send emails: ${result.message}`);
      }
    } catch (err) {
      // Catch and log any errors in the request or response
      console.error('Error sending emails:', err);
      alert('An error occurred. Please try again.');
    }
});
