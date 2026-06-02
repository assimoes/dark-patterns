DELETE FROM taxonomy_high_levels WHERE
name in (
    'Temporal Manipulation',
    'Predatory Monetization',
    'Social Exploitation',
    'Psychological Exploitation',
    'Deceptive Representation'
) and version = 1;