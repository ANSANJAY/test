SELECT DISTINCT svp_name 
FROM public.sonarqube_integration_carid_app_owners_updated
WHERE svp_name IS NOT NULL
ORDER BY svp_name;


SELECT * 
FROM public.sonarqube_integration_carid_app_owners_updated
WHERE svp_name = 'Other';
