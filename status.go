import csv
import subprocess

def get_pr_status(repo_name, pr_number):
    # Use gh CLI to fetch the pull request status
    try:
        result = subprocess.run(
            ["gh", "pr", "view", f"{pr_number}", "--repo", f"org-name/{repo_name}", "--json", "state"],
            capture_output=True,
            text=True
        )
        
        # Parse the output JSON to get the status
        if result.returncode == 0:
            status_data = result.stdout.strip()
            return status_data
        else:
            return f"Error fetching PR status: {result.stderr.strip()}"
    except Exception as e:
        return f"Exception: {str(e)}"

def process_csv(file_path):
    with open(file_path, mode='r') as file:
        csv_reader = csv.DictReader(file)
        
        # Output the status of each pull request
        for row in csv_reader:
            # Extract repo name and PR number from the pull request link
            link_parts = row['pull_request_link'].split('/')
            repo_name = link_parts[-3]
            pr_number = link_parts[-1]
            
            # Get the PR status
            pr_status = get_pr_status(repo_name, pr_number)
            print(f"Repo: {repo_name}, PR #{pr_number} - Status: {pr_status}")

# Run the script with your CSV file path
process_csv("path/to/your_file.csv")