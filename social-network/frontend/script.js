// Функция для получения и отображения пользователей
async function getUsers() {
    const response = await fetch('http://127.0.0.1:8080/users');
    const users = await response.json();
    const userList = document.getElementById('user-list');
    userList.innerHTML = ''; // Очистить старые данные
    users.forEach(user => {
        const li = document.createElement('li');
        li.textContent = `${user.username} (${user.email})`;
        userList.appendChild(li);
    });
}

// Функция для получения и отображения постов
async function getPosts() {
    const response = await fetch('http://127.0.0.1:8080/posts');
    const posts = await response.json();
    const postList = document.getElementById('post-list');
    postList.innerHTML = ''; // Очистить старые данные
    posts.forEach(post => {
        const div = document.createElement('div');
        div.classList.add('post');
        div.innerHTML = `
            <strong>${post.username}</strong>: ${post.content}<br>
            <small>Created: ${post.created_at}</small>
            <button onclick="editPost(${post.id}, prompt('New content:', '${post.content}'))">Edit</button>
            <button onclick="deletePost(${post.id})">Delete</button>
            <div id="comments-${post.id}" class="comments-section"></div>
            <textarea id="comment-${post.id}" placeholder="Write a comment..."></textarea>
            <button onclick="createComment(event, ${post.id})">Add Comment</button>
        `;
        postList.appendChild(div);
        getComments(post.id); // Загружаем комментарии для поста
    });
}

// Функция для получения и отображения комментариев
async function getComments(postId) {
    const response = await fetch('http://127.0.0.1:8080/comments');
    const comments = await response.json();
    const commentList = document.getElementById(`comments-${postId}`);
    commentList.innerHTML = ''; // Очистить старые комментарии
    comments.filter(comment => comment.post_id === postId).forEach(comment => {
        const div = document.createElement('div');
        div.classList.add('comment');
        div.innerHTML = `
            <strong>${comment.username}</strong>: ${comment.content}
            <button onclick="editComment(${comment.id}, '${comment.content}')">Edit</button>
            <button onclick="deleteComment(${comment.id})">Delete</button>
        `;
        commentList.appendChild(div);
    });
}

// Функция для создания нового поста
async function createPost(event) {
    event.preventDefault();
    const content = document.getElementById('post-content').value;
    const post = { content };
    const response = await fetch('http://127.0.0.1:8080/posts/create', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(post),
    });
    const newPost = await response.json();
    getPosts(); // Обновим список постов
}

// Функция для создания нового комментария
async function createComment(event, postId) {
    event.preventDefault();
    const textarea = document.getElementById(`comment-${postId}`);
    const content = textarea.value.trim();

    if (!content) {
        alert('Comment content cannot be empty!');
        return;
    }

    const comment = { post_id: postId, content };
    console.log('Sending comment:', comment); // Логируем данные перед отправкой

    const response = await fetch('http://127.0.0.1:8080/comments/create', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(comment),
    });

    if (response.ok) {
        textarea.value = ''; // Очистить поле комментария
        console.log('Comment added successfully!');
        getPosts(); // Обновить список постов
    } else {
        const errorText = await response.text();
        alert(`Failed to add comment: ${errorText}`);
    }
}

// Функция для редактирования поста
async function editPost(postId, newContent) {
    const response = await fetch('http://127.0.0.1:8080/posts/update', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id: postId, content: newContent }),
    });
    if (response.ok) {
        getPosts(); // Обновить список постов
    }
}

// Функция для удаления поста
async function deletePost(postId) {
    const response = await fetch('http://127.0.0.1:8080/posts/delete', {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id: postId }),
    });
    if (response.ok) {
        getPosts(); // Обновить список постов
    }
}

// Функция для редактирования комментария
async function editComment(commentId, currentContent) {
    const newContent = prompt('Edit your comment:', currentContent);
    if (newContent && newContent.trim() !== '') {
        const response = await fetch('http://127.0.0.1:8080/comments/update', {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ id: commentId, content: newContent.trim() }),
        });
        if (response.ok) {
            getPosts();
        } else {
            alert('Failed to edit comment');
        }
    }
}

// Функция для удаления комментария
async function deleteComment(commentId) {
    const confirmDelete = confirm('Are you sure you want to delete this comment?');
    if (confirmDelete) {
        const response = await fetch('http://127.0.0.1:8080/comments/delete', {
            method: 'DELETE',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ id: commentId }),
        });
        if (response.ok) {
            getPosts();
        } else {
            alert('Failed to delete comment');
        }
    }
}

// Инициализация страницы
getUsers();
getPosts();

// Подключаем события для форм
document.getElementById('create-post-form').addEventListener('submit', createPost);
