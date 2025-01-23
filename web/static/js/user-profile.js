document.addEventListener("DOMContentLoaded", () => {
    const profileView = document.getElementById("profileView")
    const profileEdit = document.getElementById("profileEdit")
    const editProfileBtn = document.getElementById("editProfile")
    const saveProfileBtn = document.getElementById("saveProfile")
    const cancelEditBtn = document.getElementById("cancelEdit")
    const messageDiv = document.getElementById("message")

    let currentUser

    // Retrieve user data from localStorage
    const storedUser = localStorage.getItem("currentUser")
    if (storedUser) {
        currentUser = JSON.parse(storedUser)
        console.log("Stored user:", currentUser)
    } else {
        console.error("No user data found in localStorage")
        showMessage("Error: User data not found", true)
        return
    }

    // Fetch and display user data
    fetchUserData()
    fetchUserPosts()

    // Event listeners
    editProfileBtn.addEventListener("click", showEditForm)
    saveProfileBtn.addEventListener("click", saveProfile)
    cancelEditBtn.addEventListener("click", cancelEdit)

    function fetchUserData() {
        fetch(`/api/user-profile/data?id=${currentUser.id}`, {
            method: "GET",
            headers: {
                Authorization: `Bearer ${currentUser.token}`,
            },
        })
            .then((response) => {
                if (!response.ok) {
                    throw new Error("Failed to fetch user data")
                }
                return response.json()
            })
            .then((data) => {
                displayUserData(data)
            })
            .catch((error) => {
                console.error("Error:", error)
                showMessage(error.message, true)
            })
    }

    function displayUserData(data) {
        document.getElementById("username").textContent = data.username
        document.getElementById("email").textContent = data.email
        document.getElementById("role").textContent = data.is_admin ? "Admin" : "User"
        document.getElementById("editUsername").value = data.username
    }

    function showEditForm() {
        profileView.style.display = "none"
        profileEdit.style.display = "block"
    }

    function saveProfile() {
        const newUsername = document.getElementById("editUsername").value
        const newPassword = document.getElementById("editPassword").value

        if (!newUsername) {
            showMessage("Username cannot be empty", true)
            return
        }

        const payload = {
            id: currentUser.id,
            username: newUsername,
            password: newPassword,
        }


        fetch("/api/user-profile/edit", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${currentUser.token}`,
            },
            body: JSON.stringify(payload),
        })
            .then((response) => {
                if (!response.ok) {
                    throw new Error("Failed to update profile")
                }
                return response.json()
            })
            .then((data) => {
                showMessage(data.message)
                fetchUserData()
                cancelEdit()
            })
            .catch((error) => {
                console.error("Error:", error)
                showMessage(error.message, true)
            })
    }

    function cancelEdit() {
        profileEdit.style.display = "none"
        profileView.style.display = "block"
        document.getElementById("editPassword").value = ""
    }

    function showMessage(message, isError = false) {
        messageDiv.textContent = message
        messageDiv.className = isError ? "error" : "success"
        setTimeout(() => {
            messageDiv.textContent = ""
            messageDiv.className = ""
        }, 5000)
    }

    function fetchUserPosts() {
        fetch(`/api/user-profile/posts?id=${currentUser.id}`, {
            method: "GET",
            headers: {
                Authorization: `Bearer ${currentUser.token}`,
            },
        })
            .then((response) => {
                if (!response.ok) {
                    throw new Error("Failed to fetch user posts")
                }
                return response.json()
            })
            .then((data) => {
                displayUserPosts(data)
            })
            .catch((error) => {
                console.error("Error:", error)
                showMessage(error.message, true)
            })
    }

    function displayUserPosts(posts) {
        const postsContainer = document.getElementById("userPosts")
        postsContainer.innerHTML = ""
        if (posts.length === 0) {
            postsContainer.innerHTML = "<p>No posts yet.</p>"
            return
        }
        posts.forEach((post) => {
            const postElement = document.createElement("div")
            postElement.className = "post"
            postElement.innerHTML = `
          <p>${post.content}</p>
          <small>Posted on: ${new Date(post.created_at).toLocaleDateString()}</small>
        `
            postsContainer.appendChild(postElement)
        })
    }
})

