
Here’s a step-by-step approach to achieve your goals. I'll break it down into the necessary SQL commands and outline how to import the CSV data.

1. Create a Copy of the Table
You can create a copy of sonar_qube_integration_github_metadata as follows:

sql
Copy code
CREATE TABLE sonar_qube_integration_github_metadata_copy AS 
SELECT * 
FROM sonar_qube_integration_github_metadata;
2. Load CSV Data into a Temporary Table
Assuming your CSV file has columns for name (repo name) and central_id (car ID), first, create a temporary table to store the CSV data:

sql
Copy code
CREATE TEMP TABLE temp_repo_car_id (
    name VARCHAR(255),
    central_id INTEGER
);
Next, load the CSV data into temp_repo_car_id. If using a PostgreSQL client like psql, you could use:

sql
Copy code
COPY temp_repo_car_id (name, central_id)
FROM '/path/to/your/file.csv'
DELIMITER ','
CSV HEADER;
Replace '/path/to/your/file.csv' with the actual file path.

3. Add the New Columns in the Copy Table
Now, add the car_updated column in the copy table to mark whether a row has been updated:

sql
Copy code
ALTER TABLE sonar_qube_integration_github_metadata_copy
ADD COLUMN car_updated BOOLEAN DEFAULT FALSE;
4. Update central_id and car_updated in the Copy Table
Now that you have the CSV data in temp_repo_car_id, update the central_id and mark car_updated as TRUE wherever a match is found:

sql
Copy code
UPDATE sonar_qube_integration_github_metadata_copy AS original
SET central_id = temp.central_id,
    car_updated = TRUE
FROM temp_repo_car_id AS temp
WHERE original.name = temp.name;
Summary
With these steps:

A copy of the original table (sonar_qube_integration_github_metadata_copy) is created.
Data from the CSV is loaded into a temporary table.
The central_id values are updated where the repository name matches, and car_updated is set to TRUE for each update.


    ------------------------


    CREATE TABLE central_id_metadata AS
SELECT DISTINCT central_id
FROM sonar_qube_integration_github_metadata_copy
WHERE central_id IS NOT NULL;

    -----------------------------------

To update the central_id_metadata table with the additional fields from your CSV based on central_id matching, follow these steps:

1. Load CSV Data into a Temporary Table
First, create a temporary table that mirrors the structure of your CSV data.

sql
Copy code
CREATE TEMP TABLE temp_central_id_update (
    central_id INTEGER,
    application_name VARCHAR(255),
    application_owner VARCHAR(255),
    application_manager VARCHAR(255),
    application_vp1 VARCHAR(255),
    application_vp2 VARCHAR(255),
    application_svp VARCHAR(255)
);
Load your CSV data into this temporary table. Using COPY is an efficient way to handle this:

sql
Copy code
COPY temp_central_id_update (central_id, application_name, application_owner, application_manager, application_vp1, application_vp2, application_svp)
FROM '/path/to/your/file.csv'
DELIMITER ','
CSV HEADER;
Replace '/path/to/your/file.csv' with the actual path to your CSV file.

2. Add New Columns to central_id_metadata
Now, add the new columns to central_id_metadata to store the additional fields from the CSV:

sql
Copy code
ALTER TABLE central_id_metadata
ADD COLUMN application_name VARCHAR(255),
ADD COLUMN application_owner VARCHAR(255),
ADD COLUMN application_manager VARCHAR(255),
ADD COLUMN application_vp1 VARCHAR(255),
ADD COLUMN application_vp2 VARCHAR(255),
ADD COLUMN application_svp VARCHAR(255);
3. Update central_id_metadata Based on central_id Matching
Use an UPDATE query to populate the new columns in central_id_metadata with data from the temporary table:

sql
Copy code
UPDATE central_id_metadata AS main
SET application_name = temp.application_name,
    application_owner = temp.application_owner,
    application_manager = temp.application_manager,
    application_vp1 = temp.application_vp1,
    application_vp2 = temp.application_vp2,
    application_svp = temp.application_svp
FROM temp_central_id_update AS temp
WHERE main.central_id = temp.central_id;
Clean Up (Optional)
After the update, you can drop the temporary table if it's no longer needed:

sql
Copy code
DROP TABLE temp_central_id_update;
This process will populate central_id_metadata with the new fields based on matching central_id values. Let me know if you need any further adjustments!
