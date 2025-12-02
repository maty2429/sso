-- ==================================================================
-- 0. LIMPIEZA TOTAL (RESET)
-- ==================================================================
DROP TRIGGER IF EXISTS set_timestamp_users ON users;
DROP FUNCTION IF EXISTS trigger_set_timestamp;
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS refresh_tokens CASCADE;
DROP TABLE IF EXISTS project_member_roles CASCADE;
DROP TABLE IF EXISTS project_members CASCADE;
DROP TABLE IF EXISTS role_definitions CASCADE;
DROP TABLE IF EXISTS projects CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- ==================================================================
-- 1. EXTENSIONES Y FUNCIONES
-- ==================================================================
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ==================================================================
-- 2. TABLA USUARIOS (Permite creación sin password)
-- ==================================================================
CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- RUT único para Login
    rut integer NOT NULL,        
    dv char(1) NOT NULL,         
    full_rut varchar(15) GENERATED ALWAYS AS (rut::text || '-' || dv) STORED,

    first_name varchar(100) NOT NULL,
    last_name varchar(100) NOT NULL,
    email varchar(150) UNIQUE NOT NULL,
    
    -- IMPORTANTE: Acepta NULL para usuarios recién creados (invitados)
    password_hash text, 
    
    -- ESTADOS DE LOGIN
    must_change_password boolean DEFAULT TRUE, -- TRUE = Forzar creación de clave
    is_active boolean DEFAULT TRUE,            -- TRUE = Usuario habilitado
    
    password_changed_at timestamp,             
    recovery_token varchar(100),
    recovery_token_expires_at timestamp,
    failed_attempts smallint DEFAULT 0,
    locked_until timestamp, 
    last_login_at timestamp,
    created_at timestamp DEFAULT NOW(),
    updated_at timestamp DEFAULT NOW(),

    CONSTRAINT uq_users_rut UNIQUE (rut),
    CONSTRAINT uq_users_full_rut UNIQUE (full_rut)
);

CREATE TRIGGER set_timestamp_users
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION trigger_set_timestamp();

CREATE INDEX idx_users_full_rut ON users(full_rut);

-- ==================================================================
-- 3. PROYECTOS (Con URL para redirección)
-- ==================================================================
CREATE TABLE projects (
    id int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    project_code varchar(50) UNIQUE NOT NULL, -- Ej: 'PRJ-CONSTRUCCION'
    name varchar(100) NOT NULL,
    
    -- NUEVO: La URL a donde el SSO debe redirigir tras el login exitoso
    frontend_url varchar(255),  -- Ej: 'https://app.construccion.cl'
    
    description text,
    is_active boolean DEFAULT TRUE, 
    created_at timestamp DEFAULT NOW(),
    updated_at timestamp DEFAULT NOW()
);

CREATE INDEX idx_projects_code ON projects(project_code);

-- ==================================================================
-- 4. DICCIONARIO DE ROLES (Numéricos)
-- ==================================================================
CREATE TABLE role_definitions (
    code smallint PRIMARY KEY, -- El número (ID real del rol)
    name varchar(50) NOT NULL, -- El nombre humano ('Admin', 'Visita')
    description text
);

-- Insertamos roles estándar
INSERT INTO role_definitions (code, name, description) VALUES 
(10, 'Visualizador', 'Solo lectura'),
(20, 'Digitador', 'Ingreso de datos básicos'),
(50, 'Supervisor', 'Aprobación y gestión de equipo'),
(99, 'Admin', 'Configuración total del proyecto');

-- ==================================================================
-- 5. MIEMBROS (Quién está en qué proyecto)
-- ==================================================================
CREATE TABLE project_members (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id int NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    
    is_active boolean DEFAULT TRUE, 
    joined_at timestamp DEFAULT NOW(),

    -- Evita duplicados: Juan solo puede ser "miembro" de un proyecto una vez.
    -- Sus múltiples roles se definen en la tabla siguiente.
    CONSTRAINT uq_member_project UNIQUE (user_id, project_id)
);

CREATE INDEX idx_members_user ON project_members(user_id);
CREATE INDEX idx_members_project ON project_members(project_id);

-- ==================================================================
-- 6. ASIGNACIÓN DE ROLES (Muchos roles por miembro)
-- ==================================================================
-- Esta tabla permite que Juan tenga rol 10 y rol 50 en el mismo proyecto
CREATE TABLE project_member_roles (
    member_id uuid NOT NULL REFERENCES project_members(id) ON DELETE CASCADE,
    role_code smallint NOT NULL REFERENCES role_definitions(code) ON DELETE CASCADE,
    assigned_at timestamp DEFAULT NOW(),
    
    PRIMARY KEY (member_id, role_code)
);

-- ==================================================================
-- 7. TABLAS DE SOPORTE (Tokens y Logs)
-- ==================================================================
CREATE TABLE refresh_tokens (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash text NOT NULL,       
    device_info varchar(255),       
    ip_address inet,
    is_revoked boolean DEFAULT FALSE, 
    expires_at timestamp NOT NULL,    
    created_at timestamp DEFAULT NOW()
);

CREATE TABLE audit_logs (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id uuid REFERENCES users(id) ON DELETE SET NULL,
    project_id int REFERENCES projects(id) ON DELETE SET NULL,
    action varchar(50) NOT NULL,  
    description text,
    ip_address inet,
    meta_data jsonb,              
    created_at timestamp DEFAULT NOW()
);
