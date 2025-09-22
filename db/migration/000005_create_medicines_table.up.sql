DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'medicine_unit') THEN
    CREATE TYPE medicine_unit AS ENUM ('tablet', 'capsule', 'box', 'bottle');
  END IF;
END$$;

CREATE TABLE medicines (
  id SERIAL PRIMARY KEY,
  name VARCHAR NOT NULL,
  unit medicine_unit NOT NULL,
  price NUMERIC(15,2) NOT NULL DEFAULT 0,
  stock INT NOT NULL DEFAULT 0,
  description TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER trg_medicines_updated_at
BEFORE UPDATE ON medicines
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- Seed dữ liệu mẫu
INSERT INTO medicines (name, unit, price, stock, description) VALUES
('Paracetamol', 'tablet', 2500, 100, 'Pain reliever and fever reducer'),
('Amoxicillin', 'capsule', 5000, 200, 'Antibiotic used to treat infections'),
('Vitamin C', 'box', 30000, 50, 'Immune system booster'),
('Cough Syrup', 'bottle', 45000, 30, 'Used to relieve cough');
