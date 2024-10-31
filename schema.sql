CREATE TABLE users (
    user_id INT UNSIGNED NOT NULL AUTO_INCREMENT,

    signup_date DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_signin_date DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    signin_locked BOOL NOT NULL DEFAULT false,
    signin_locked_date DATETIME,

    is_admin BOOL NOT NULL DEFAULT false,

    email VARCHAR(255) NOT NULL UNIQUE,
    full_name VARCHAR(255) NOT NULL,
    business_category ENUM('STUDENT', 'TEACHER') NOT NULL,
    department_number VARCHAR(255) NOT NULL,

    PRIMARY KEY (user_id)
);

CREATE TABLE sessions (
    session_id INT UNSIGNED NOT NULL AUTO_INCREMENT,

    user_id INT UNSIGNED NOT NULL,
    creation_date DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    session_token VARCHAR(255) NOT NULL UNIQUE,

    PRIMARY KEY (session_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id)
);

CREATE TABLE events (
    event_id INT UNSIGNED NOT NULL AUTO_INCREMENT,

    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    event_date DATETIME NOT NULL,
    creation_date DATETIME DEFAULT CURRENT_TIMESTAMP,

    parent_event_id INT UNSIGNED,

    PRIMARY KEY (event_id),
    FOREIGN KEY (parent_event_id) REFERENCES events(event_id)
);

CREATE TABLE photos (
    photo_id INT UNSIGNED NOT NULL AUTO_INCREMENT,

    path_to_photo VARCHAR(255) NOT NULL,
    creation_date DATETIME DEFAULT CURRENT_TIMESTAMP,

    event_id INT UNSIGNED NOT NULL,

    PRIMARY KEY (photo_id),
    FOREIGN KEY (event_id) REFERENCES events(event_id)
);

CREATE TABLE user_folders (
    user_folder_id INT UNSIGNED NOT NULL AUTO_INCREMENT,

    is_sub_folder BOOL DEFAULT false,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    creation_date DATETIME DEFAULT CURRENT_TIMESTAMP,

    user_id INT UNSIGNED NOT NULL,
    parent_folder_id INT UNSIGNED,

    PRIMARY KEY (folder_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id),
    FOREIGN KEY (parent_folder_id) REFERENCES user_folders(folder_id)
);

CREATE TABLE recognized_users (
    recognized_user_id INT UNSIGNED NOT NULL AUTO_INCREMENT,

    user_id INT UNSIGNED NOT NULL,
    photo_id INT UNSIGNED NOT NULL,

    PRIMARY KEY (id),
    FOREIGN KEY (user_id) REFERENCES users(user_id),
    FOREIGN KEY (photo_id) REFERENCES photos(photo_id)
);
