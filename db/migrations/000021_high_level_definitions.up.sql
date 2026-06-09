ALTER TABLE taxonomy_high_levels ADD COLUMN definition text;

UPDATE taxonomy_high_levels SET definition =
    $$Strategies that exploit, distort, or weaponize the player's time investment to serve the game creator's goals (retention, monetization) at the expense of the player's autonomy over how they spend their time.$$
    WHERE name = 'Temporal Manipulation';
UPDATE taxonomy_high_levels SET definition =
    $$Strategies that extract money from players through coercive, deceptive, or exploitative purchasing systems that obscure true costs, exploit psychological vulnerabilities, or create artificial need.$$
    WHERE name = 'Predatory Monetization';
UPDATE taxonomy_high_levels SET definition =
    $$Strategies that weaponize the player's social relationships, social identity, or need for social belonging to drive engagement, recruitment, or spending.$$
    WHERE name = 'Social Exploitation';
UPDATE taxonomy_high_levels SET definition =
    $$Strategies that exploit known cognitive biases, emotional vulnerabilities, or behavioral psychology principles to manipulate player decisions about time, money, or engagement.$$
    WHERE name = 'Psychological Exploitation';
UPDATE taxonomy_high_levels SET definition =
    $$Strategies that mislead the player about what the game is, what they are purchasing, or what they will receive — creating a gap between expectation and reality.$$
    WHERE name = 'Deceptive Representation';
