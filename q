SELECT DISTINCT svp_name 
FROM public.sonarqube_integration_carid_app_owners_updated
WHERE svp_name IS NOT NULL
ORDER BY svp_name;


SELECT * 
FROM public.sonarqube_integration_carid_app_owners_updated
WHERE svp_name = 'Other';


-- Step 1: Get distinct car_id under each SVP
SELECT 
    o.svp_name, 
    ARRAY_AGG(DISTINCT o.car_id) AS car_ids
FROM 
    public.sonarqube_integration_carid_app_owners_updated o
GROUP BY 
    o.svp_name
ORDER BY 
    o.svp_name;

-- Step 2: Get repository names under each SVP
SELECT 
    o.svp_name, 
    ARRAY_AGG(DISTINCT r.name) AS repo_names
FROM 
    public.sonar_qube_integration_github_metadata_central_id r
JOIN 
    public.sonarqube_integration_carid_app_owners_updated o
    ON r.central_id = o.car_id
GROUP BY 
    o.svp_name
ORDER BY 
    o.svp_name;

-- Step 3: Get active repositories under each SVP (last 5 months)
SELECT 
    o.svp_name, 
    ARRAY_AGG(DISTINCT r.name) AS active_repo_names
FROM 
    public.sonar_qube_integration_github_metadata_central_id r
JOIN 
    public.sonarqube_integration_carid_app_owners_updated o
    ON r.central_id = o.car_id
WHERE 
    r.pushed_at >= '2024-05-16'
GROUP BY 
    o.svp_name
ORDER BY 
    o.svp_name;


------------------------
1. Count of Distinct Car IDs under Each SVP:


SELECT 
    o.svp_name, 
    COUNT(DISTINCT o.car_id) AS car_id_count
FROM 
    public.sonarqube_integration_carid_app_owners_updated o
GROUP BY 
    o.svp_name
ORDER BY 
    car_id_count DESC;
------------------------

------------------------
2. Count of Repositories under Each SVP:


SELECT 
    o.svp_name, 
    COUNT(DISTINCT r.name) AS repo_count
FROM 
    public.sonar_qube_integration_github_metadata_central_id r
JOIN 
    public.sonarqube_integration_carid_app_owners_updated o
    ON r.central_id = o.car_id
GROUP BY 
    o.svp_name
ORDER BY 
    repo_count DESC;

------------------------

3. Count of Active Repositories under Each SVP:

SELECT 
    o.svp_name, 
    COUNT(DISTINCT r.name) AS active_repo_count
FROM 
    public.sonar_qube_integration_github_metadata_central_id r
JOIN 
    public.sonarqube_integration_carid_app_owners_updated o
    ON r.central_id = o.car_id
WHERE 
    r.pushed_at >= '2024-05-16'
GROUP BY 
    o.svp_name
ORDER BY 
    active_repo_count DESC;



