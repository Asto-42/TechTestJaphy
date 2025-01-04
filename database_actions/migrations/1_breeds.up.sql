CREATE TABLE IF NOT EXISTS core.breeds (
    id INT AUTO_INCREMENT PRIMARY KEY,
    species VARCHAR(50) NOT NULL,
    pet_size VARCHAR(50) NOT NULL,
    name VARCHAR(500) NOT NULL,
    weight_min INT NOT NULL,
    weight_max INT NOT NULL
);
