CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL
);

CREATE TABLE employees (
    id SERIAL PRIMARY KEY,                
    first_name VARCHAR(100) NOT NULL,     
    last_name VARCHAR(100) NOT NULL,      
    email VARCHAR(255) NOT NULL UNIQUE,   
    age INT,             
    created_at TIMESTAMP DEFAULT NOW()    
);

CREATE TABLE sessions (
    token CHAR(36) PRIMARY KEY NOT NULL,
    user_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    last_used TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE books (
    id INT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    author VARCHAR(255) NOT NULL
);


INSERT INTO employees (first_name, last_name, email, age)
VALUES ('John', 'Doe', 'john.doe@example.com', 30);

INSERT INTO employees (first_name, last_name, email, age)
VALUES ('Jane', 'Smith', 'jane.smith@example.com', 25);

INSERT INTO employees (first_name, last_name, email, age)
VALUES ('Alice', 'Johnson', 'alice.johnson@example.com', 40);

INSERT INTO employees (first_name, last_name, email, age)
VALUES ('Bob', 'Williams', 'bob.williams@example.com', 35);

INSERT INTO employees (first_name, last_name, email, age)
VALUES ('Charlie', 'Brown', 'charlie.brown@example.com', 28);


INSERT INTO books (id, title, author) VALUES
(1, 'To Kill a Mockingbird', 'Harper Lee'),
(2, '1984', 'George Orwell'),
(3, 'Pride and Prejudice', 'Jane Austen'),
(4, 'The Great Gatsby', 'F. Scott Fitzgerald'),
(5, 'Moby Dick', 'Herman Melville');
