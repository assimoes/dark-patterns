DELETE FROM taxonomy_meso_levels
WHERE code IN (
    'TM-1','TM-2','TM-3','TM-4',
    'PM-1','PM-2','PM-3','PM-4','PM-5','PM-6',
    'SE-1','SE-2','SE-3',
    'PE-1','PE-2','PE-3','PE-4',
    'DR-1','DR-2'
) AND version = 1;