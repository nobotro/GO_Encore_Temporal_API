CREATE TABLE bills
(
    id TEXT,
    status  VARCHAR DEFAULT 'open',
    created_at DATE NOT NULL DEFAULT CURRENT_DATE,
    closed_at DATE,
    currency TEXT NOT NULL,
    total_amount DOUBLE PRECISION NOT NULL DEFAULT 0,
    PRIMARY KEY(id)

);


CREATE TABLE lineitems
(
    id integer generated always as identity,
    bill_id TEXT NOT NULL,
    itemdate DATE  DEFAULT CURRENT_DATE,
    amount DOUBLE PRECISION NOT NULL DEFAULT 0,
    PRIMARY KEY(id),
    CONSTRAINT fk_bill
        FOREIGN KEY(bill_id)
            REFERENCES bills(id)
            ON DELETE SET NULL

);