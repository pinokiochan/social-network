let currentUser = null;
let currentPage = 1;
const pageSize = 10;

function showAuthForms() {
    document.getElementById('auth-forms').style.display = 'block';
    document.getElementById('content').style.display = 'none';
}

function showContent() {
    document.getElementById('auth-forms').style.display = 'none';
    document.getElementById('content').style.display = 'block';
}

async function register(event) {
    event.preventDefault();
    const username = document.getElementById('register-username').value;
    const email = document.getElementById('register-email').value;
    const password = document.getElementById('register-password').value;

    try {
        const response = await fetch('https://social-network-2.onrender.com/api/register', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, email, password })
        });
        if (response.ok) {
            alert('Registration successful. Please log in.');
        } else {
            alert('Registration failed. Please try again.');
        }
    } catch (error) {
        console.error('Error:', error);
    }
}

async function login(event) {
    event.preventDefault();
    const email = document.getElementById('login-email').value;
    const password = document.getElementById('login-password').value;

    try {
        const response = await fetch('https://social-network-2.onrender.com/api/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password })
        });
        if (response.ok) {
            const data = await response.json();
            currentUser = { id: data.user_id, email: email };
            localStorage.setItem('token', data.token);
            localStorage.setItem('currentUser', JSON.stringify(currentUser));
            console.log('Logged in user:', currentUser);
            showContent();
            getPosts();
        } else {
            alert('Login failed. Please try again.');
        }
    } catch (error) {
        console.error('Error:', error);
    }
}

function logout() {
    localStorage.removeItem('token');
    localStorage.removeItem('currentUser');
    currentUser = null;
    showAuthForms();
}

async function getPosts(searchParams = {}) {
    const token = localStorage.getItem('token');
    if (!token) {
        console.error('No token found');
        showAuthForms();
        return;
    }

    try {
        const queryParams = new URLSearchParams({
            page: searchParams.page || currentPage,
            page_size: pageSize,
            ...searchParams
        });

        const response = await fetch(`https://social-network-2.onrender.com/api/posts?${queryParams}`, {
            headers: { 'Authorization': token }
        });
        if (!response.ok) {
            throw new Error('Failed to fetch posts');
        }
        const posts = await response.json();
        const postList = document.getElementById('post-list');
        postList.innerHTML = '';
        posts.forEach(post => {
            const div = document.createElement('div');
            div.classList.add('post');
            div.innerHTML = `
                <strong>${post.username}</strong>: ${post.content}<br>
                <small>Created: ${formatDate(post.created_at)}</small>
                <div class="post-actions">
                    ${post.user_id === currentUser.id ? `
                        <button onclick="editPost(${post.id}, '${post.content.replace(/'/g, "\\'")}')" class="edit-btn">
                            <i class="fas fa-edit"></i> 
                        </button>
                        <button onclick="deletePost(${post.id})" class="delete-btn">
                            <i class="fas fa-trash-alt"></i> 
                        </button>
                    ` : ''}
                </div>
                <div id="comments-${post.id}" class="comments-section"></div>
                <textarea id="comment-${post.id}" placeholder="Write a comment..."></textarea>
                <button class="add-comment-btn" data-post-id="${post.id}">
                    <i class="fas fa-comment"></i> Add Comment
                </button>
            `;
            postList.appendChild(div);
            getComments(post.id);
        });

        updatePagination();
    } catch (error) {
        console.error('Error fetching posts:', error);
    }
}


function updatePagination() {
    const paginationContainer = document.getElementById('pagination');
    paginationContainer.innerHTML = `
        <button onclick="changePage(${currentPage - 1})" ${currentPage === 1 ? 'disabled' : ''}>Previous</button>
        <span>Page ${currentPage}</span>
        <button onclick="changePage(${currentPage + 1})">Next</button>
    `;
}

function changePage(newPage) {
    if (newPage < 1) return;
    currentPage = newPage;
    getPosts();
}

function formatDate(isoDate) {
    const date = new Date(isoDate);
    const options = {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
        hour: 'numeric',
        minute: 'numeric',
        hour12: true
    };
    return date.toLocaleString('en-US', options);
}

async function searchPosts() {
    const keyword = document.getElementById('search-input').value;
    const userId = document.getElementById('user-filter').value;
    const date = document.getElementById('date-filter').value;

    const searchParams = {};
    if (keyword) searchParams.keyword = keyword;
    if (userId) searchParams.user_id = userId;
    if (date) searchParams.date = date;

    currentPage = 1;
    await getPosts(searchParams);
}

async function createPost(event) {
    event.preventDefault();
    const content = document.getElementById('post-content').value;
    const token = localStorage.getItem('token');
    if (!token) {
        console.error('No token found');
        showAuthForms();
        return;
    }

    try {
        const response = await fetch('https://social-network-2.onrender.com/posts/create', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': token
            },
            body: JSON.stringify({ content }),
        });
        if (!response.ok) {
            throw new Error('Failed to create post');
        }
        const newPost = await response.json();
        getPosts();
    } catch (error) {
        console.error('Error creating post:', error);
    }
}

async function editPost(postId, currentContent) {
    const newContent = prompt('Edit your post:', currentContent);
    if (newContent !== null && newContent.trim() !== '') {
        const token = localStorage.getItem('token');
        if (!token) {
            console.error('No token found');
            showAuthForms();
            return;
        }

        try {
            const response = await fetch('https://social-network-2.onrender.com/api/posts/update', {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': token
                },
                body: JSON.stringify({ id: postId, content: newContent.trim() }),
            });
            if (!response.ok) {
                throw new Error('Failed to edit post');
            }
            getPosts();
        } catch (error) {
            console.error('Error editing post:', error);
            alert('Failed to edit post. Please try again.');
        }
    }
}

async function deletePost(postId) {
    const token = localStorage.getItem('token');
    if (!token) {
        console.error('No token found');
        showAuthForms();
        return;
    }

    try {
        const response = await fetch('https://social-network-2.onrender.com/api/posts/delete', {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': token
            },
            body: JSON.stringify({ id: postId }),
        });
        if (!response.ok) {
            throw new Error('Failed to delete post');
        }
        getPosts();
    } catch (error) {
        console.error('Error deleting post:', error);
        alert('Failed to delete post. Please try again.');
    }
}

async function getComments(postId) {
    const token = localStorage.getItem('token');
    if (!token) {
        console.error('No token found');
        return;
    }

    try {
        const response = await fetch('https://social-network-2.onrender.com/api/comments', {
            headers: { 'Authorization': token }
        });
        if (!response.ok) {
            throw new Error('Failed to fetch comments');
        }
        const comments = await response.json();
        const commentList = document.getElementById(`comments-${postId}`);
        commentList.innerHTML = '';
        comments.filter(comment => comment.post_id === postId).forEach(comment => {
            const div = document.createElement('div');
            div.classList.add('comment');
            div.innerHTML = `
                <strong>${comment.username}</strong>: ${comment.content}
                <div class="comment-actions">
                    ${comment.user_id === currentUser.id ? `
                        <button onclick="editComment(${comment.id}, '${comment.content.replace(/'/g, "\\'")}')" class="edit-btn">
                            <i class="fas fa-edit"></i> 
                        </button>
                        <button onclick="deleteComment(${comment.id})" class="delete-btn">
                            <i class="fas fa-trash-alt"></i> 
                        </button>
                    ` : ''}
                </div>
            `;
            commentList.appendChild(div);
        });
    } catch (error) {
        console.error('Error fetching comments:', error);
    }
}


async function createComment(event, postId) {
    event.preventDefault();
    const textarea = document.getElementById(`comment-${postId}`);
    const content = textarea.value.trim();

    if (!content) {
        alert('Comment content cannot be empty!');
        return;
    }

    const token = localStorage.getItem('token');
    if (!token) {
        console.error('No token found');
        showAuthForms();
        return;
    }

    try {
        const response = await fetch('https://social-network-2.onrender.com/api/comments/create', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': token
            },
            body: JSON.stringify({ post_id: postId, content }),
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const result = await response.json();
        console.log('Comment created:', result);
        textarea.value = '';
        await getComments(postId);
    } catch (error) {
        console.error('Error creating comment:', error);
        alert('Failed to create comment. Please try again.');
    }
}

async function editComment(commentId, currentContent) {
    const newContent = prompt('Edit your comment:', currentContent);
    if (newContent !== null && newContent.trim() !== '') {
        const token = localStorage.getItem('token');
        if (!token) {
            console.error('No token found');
            showAuthForms();
            return;
        }

        try {
            const response = await fetch('https://social-network-2.onrender.com/api/comments/update', {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': token
                },
                body: JSON.stringify({ id: commentId, content: newContent.trim() }),
            });
            if (!response.ok) {
                throw new Error('Failed to edit comment');
            }
            getPosts();
        } catch (error) {
            console.error('Error editing comment:', error);
            alert('Failed to edit comment. Please try again.');
        }
    }
}

async function deleteComment(commentId) {
    const token = localStorage.getItem('token');
    if (!token) {
        console.error('No token found');
        showAuthForms();
        return;
    }

    try {
        const response = await fetch('https://social-network-2.onrender.com/api/comments/delete', {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': token
            },
            body: JSON.stringify({ id: commentId }),
        });
        if (!response.ok) {
            throw new Error('Failed to delete comment');
        }
        getPosts();
    } catch (error) {
        console.error('Error deleting comment:', error);
        alert('Failed to delete comment. Please try again.');
    }
}

document.addEventListener('DOMContentLoaded', () => {
    const token = localStorage.getItem('token');
    const storedUser = localStorage.getItem('currentUser');
    if (token && storedUser) {
        currentUser = JSON.parse(storedUser);
        console.log('Stored user:', currentUser);
        showContent();
        getPosts();
    } else {
        showAuthForms();
    }

    document.getElementById('register-form').addEventListener('submit', register);
    document.getElementById('login-form').addEventListener('submit', login);
    document.getElementById('logout-btn').addEventListener('click', logout);
    document.getElementById('create-post-form').addEventListener('submit', createPost);
    document.getElementById('search-form').addEventListener('submit', (e) => {
        e.preventDefault();
        searchPosts();
    });

    document.addEventListener('click', (event) => {
        if (event.target.classList.contains('add-comment-btn')) {
            const postId = parseInt(event.target.getAttribute('data-post-id'));
            createComment(event, postId);
        }
    });
});