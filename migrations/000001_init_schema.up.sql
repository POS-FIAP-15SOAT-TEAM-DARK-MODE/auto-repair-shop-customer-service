-- Create USER table
CREATE TABLE "user" (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create ROLE table
CREATE TABLE "role" (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    created_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create USER_ROLE join table
CREATE TABLE user_role (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES "user" (id) ON DELETE CASCADE,
    role_id VARCHAR(36) NOT NULL REFERENCES "role" (id) ON DELETE CASCADE,
    UNIQUE (user_id, role_id)
);

-- Create CUSTOMER table
CREATE TABLE customer (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL UNIQUE REFERENCES "user" (id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL,
    cpf VARCHAR(14) UNIQUE,
    cnpj VARCHAR(18) UNIQUE,
    company_name VARCHAR(255),
    phone VARCHAR(25) NOT NULL,
    created_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        CHECK (
            (
                type = 'INDIVIDUAL'
                AND cpf IS NOT NULL
                AND cnpj IS NULL
                AND company_name IS NULL
            )
            OR (
                type = 'COMPANY'
                AND cpf IS NULL
                AND cnpj IS NOT NULL
                AND company_name IS NOT NULL
            )
        )
);

-- Create VEHICLE table
CREATE TABLE vehicle (
    id VARCHAR(36) PRIMARY KEY,
    license_plate VARCHAR(20) UNIQUE NOT NULL,
    brand VARCHAR(100) NOT NULL,
    model VARCHAR(100) NOT NULL,
    year INT NOT NULL,
    customer_id VARCHAR(36) NOT NULL REFERENCES customer (id) ON DELETE CASCADE,
    created_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for foreign keys and search columns
CREATE INDEX idx_customer_user_id ON customer (user_id);

CREATE INDEX idx_vehicle_customer_id ON vehicle (customer_id);

CREATE INDEX idx_user_role_user_id ON user_role (user_id);

CREATE INDEX idx_user_role_role_id ON user_role (role_id);

-- Insert default roles
INSERT INTO "role" (id, name) VALUES ('1', 'ADMIN');

INSERT INTO "role" (id, name) VALUES ('2', 'MECHANIC');

INSERT INTO "role" (id, name) VALUES ('3', 'ATTENDANT');

INSERT INTO "role" (id, name) VALUES ('4', 'CUSTOMER');
