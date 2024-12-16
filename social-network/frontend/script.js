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
            <div id="comments-${post.id}"></div>
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
        div.innerHTML = `<strong>${comment.username}</strong>: ${comment.content}`;
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
    const content = document.getElementById(`comment-${postId}`).value;
    const comment = { post_id: postId, content };
    const response = await fetch('http://127.0.0.1:8080/comments/create', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(comment),
    });
    const newComment = await response.json();
    getComments(postId); // Обновим комментарии для поста
}

// Инициализация страницы
getUsers();
getPosts();

// Подключаем события для форм
document.getElementById('create-post-form').addEventListener('submit', createPost);
