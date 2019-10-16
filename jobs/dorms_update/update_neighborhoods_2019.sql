INSERT INTO neighborhoods (created_at, updated_at, name, trakked)
VALUES
       (NOW(), NOW(), 'East Campus', 1),
       (NOW(), NOW(), 'West Campus', 1),
       (NOW(), NOW(), 'North Campus', 1),
       (NOW(), NOW(), 'Central Campus', 1);

UPDATE dorms SET neighborhood_id=(
    SELECT id FROM neighborhoods WHERE name='East Campus'
    )
WHERE name IN (
               'Currier', 'East', 'Fayerweather', 'Fitch', 'Prospect'
              );

UPDATE dorms SET neighborhood_id=(
    SELECT id FROM neighborhoods WHERE name='West Campus'
)
WHERE name IN (
               'Bryant', 'Carter', 'Garfield', 'Gladden', 'Mark Hopkins', 'Perry', 'Wood'
    );

UPDATE dorms SET neighborhood_id=(
    SELECT id FROM neighborhoods WHERE name='North Campus'
)
WHERE name IN (
               'Dodd', 'Goodrich', 'Hubbell', 'Lehman', 'Parsons', 'Sewall', 'Thompson', 'Tyler', 'Tyler Annex'
    );

UPDATE dorms SET neighborhood_id=(
    SELECT id FROM neighborhoods WHERE name='Central Campus'
)
WHERE name IN (
               'Agard', 'Brooks', 'Horn', 'Morgan', 'Spencer', 'West'
    );

UPDATE neighborhoods SET neighborhoods.deleted_at=NOW() WHERE name in ('Currier', 'Wood', 'Spencer', 'Dodd');