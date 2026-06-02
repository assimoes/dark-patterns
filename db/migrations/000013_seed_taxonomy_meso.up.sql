INSERT INTO taxonomy_meso_levels (parent_id, code, name, description, version)
SELECT p.id, v.code, v.name, v.description, 1
FROM (VALUES
    ('Temporal Manipulation', 'TM-1', 'Artificial Extension', 'padding playtime to stretch engagement'),
    ('Temporal Manipulation', 'TM-2', 'Scheduled Coercion', 'time-gated demands that punish absence'),
    ('Temporal Manipulation', 'TM-3', 'Session Manipulation', 'mechanics that prolong or trap a session'),
    ('Temporal Manipulation', 'TM-4', 'Forced Ad Exposure', 'unskippable or coerced advertising'),
    ('Predatory Monetization', 'PM-1', 'Pay-to-Progress', 'paying to advance or to win'),
    ('Predatory Monetization', 'PM-2', 'Currency Obfuscation', 'hiding real cost behind premium currencies'),
    ('Predatory Monetization', 'PM-3', 'Gambling Mechanics', 'loot boxes and chance-based purchases'),
    ('Predatory Monetization', 'PM-4', 'Price Manipulation', 'misleading or shifting pricing'),
    ('Predatory Monetization', 'PM-5', 'Recurring/Compounding Charges', 'subscriptions and stacking charges'),
    ('Predatory Monetization', 'PM-6', 'Interface Monetization Traps', 'UI nudges towards unintended spending'),
    ('Social Exploitation', 'SE-1', 'Recruitment Pressure', 'pressure to bring others in'),
    ('Social Exploitation', 'SE-2', 'Social Obligation', 'guilt or duty towards other players'),
    ('Social Exploitation', 'SE-3', 'Competitive Pressure', 'leveraging rivalry to drive behaviour'),
    ('Psychological Exploitation', 'PE-1', 'Loss Aversion Exploitation', 'exploiting fear of losing progress'),
    ('Psychological Exploitation', 'PE-2', 'Completionism Exploitation', 'exploiting the urge to complete'),
    ('Psychological Exploitation', 'PE-3', 'Cognitive Bias Exploitation', 'exploiting general cognitive biases'),
    ('Psychological Exploitation', 'PE-4', 'Sensory Manipulation', 'audiovisual manipulation of attention'),
    ('Deceptive Representation', 'DR-1', 'Misleading Marketing', 'pre-purchase misrepresentation'),
    ('Deceptive Representation', 'DR-2', 'Post-Purchase Deception', 'misrepresentation after the sale')
) AS v(family, code, name, description)
JOIN taxonomy_high_levels p ON p.name = v.family
ON CONFLICT (code) DO NOTHING;