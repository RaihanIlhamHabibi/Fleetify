-- Fleetify seed data (schema loaded from 01-schema.sql)
INSERT INTO users (username, role) VALUES
    ('advisor_sa', 'SA'),
    ('manager_approval', 'APPROVAL');

INSERT INTO vehicles (license_plate, model) VALUES
    ('B 1234 XYZ', 'Toyota Avanza'),
    ('B 5678 ABC', 'Honda Brio'),
    ('B 9012 DEF', 'Mitsubishi Xpander');

INSERT INTO master_items (item_name, type, price) VALUES
    ('Oli Mesin 1L', 'PART', 85000.00),
    ('Filter Oli', 'PART', 45000.00),
    ('Kampas Rem Depan', 'PART', 320000.00),
    ('Jasa Ganti Oli', 'SERVICE', 75000.00),
    ('Jasa Service Berkala', 'SERVICE', 250000.00);
