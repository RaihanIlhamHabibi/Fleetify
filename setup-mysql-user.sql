-- Jalankan di phpMyAdmin (tab SQL) saat sudah login
-- Lalu sesuaikan backend\.env dengan user/password ini

CREATE DATABASE IF NOT EXISTS fleetify CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

CREATE USER IF NOT EXISTS 'fleetify'@'localhost' IDENTIFIED BY 'fleetify123';
GRANT ALL PRIVILEGES ON fleetify.* TO 'fleetify'@'localhost';
FLUSH PRIVILEGES;
