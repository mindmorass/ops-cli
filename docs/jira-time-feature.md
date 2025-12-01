# Jira Time Command

## Overview

The `jira time` command is a new addition to the Jira module that allows users to find their assigned issues and interactively log time worked on them.

## Features

- **Issue Discovery**: Automatically finds issues assigned to the current user
- **Smart Filtering**: Defaults to active sprint issues for focused time tracking
- **Interactive Selection**: Multi-select interface for choosing which issues to log time on
- **Flexible Time Input**: Supports Jira time format (h, d, m, w combinations)
- **Work Description**: Optional comments for worklog entries
- **Date Selection**: Ability to specify work date (defaults to today)
- **Batch Operations**: Log the same time period across multiple issues
- **Validation**: Input validation for time format and dates

## Usage

```bash
# Find assigned issues and log time interactively
ops-cli jira time

# Limit to 10 most recent issues
ops-cli jira time --limit 10

# Include all issues (not just active sprint)
ops-cli jira time --no-active-sprint

# Show help
ops-cli jira time --help
```

## Interactive Flow

### 1. Issue Discovery

The command searches for issues using this JQL:

```
assignee = currentUser() AND resolution = Unresolved AND sprint in openSprints() ORDER BY updated DESC
```

### 2. Issue Selection

Users can select multiple issues using a toggle interface:

```
Found 3 assigned issue(s):

1. SRE-150 - Update cdkt and alerting for ONLY production
   Type: Story | Status: In Progress

2. SRE-146 - [NR][Ansible] Define structure for config mgmt
   Type: Story | Status: In Progress

Options:
• Enter issue number(s) to toggle selection
• 'show' - show current selection
• 'clear' - clear all selections
• 'done' - proceed with selected issues
```

### 3. Time Entry

Support for Jira's flexible time format:

```
Enter time worked (Jira format):
Examples: 2h, 1d, 30m, 1h 30m, 2d 4h
• m = minutes, h = hours, d = days, w = weeks
Time spent: 2h 30m
```

### 4. Work Description (Optional)

```
Enter work description (optional):
Description: Completed infrastructure updates and testing
```

### 5. Work Date (Optional)

```
Enter work date (optional, default: today):
Format: YYYY-MM-DD or leave empty for today
Work date: 2024-10-28
```

### 6. Batch Logging

The same time and description is applied to all selected issues:

```
⏳ Logging 2h 30m of work...
✅ Logged time for SRE-150
✅ Logged time for SRE-146
✨ Time logging completed!
```

## Command Options

| Option               | Description                       | Default |
| -------------------- | --------------------------------- | ------- |
| `[max-issues]`       | Maximum number of issues to fetch | 20      |
| `--active-sprint`    | Include only active sprint issues | true    |
| `--no-active-sprint` | Include all unresolved issues     | false   |
| `--color`            | Enable colored output             | true    |
| `--no-color`         | Disable colored output            | false   |

## Time Format Support

The command supports Jira's native time format:

- **Minutes**: `30m`, `45m`
- **Hours**: `2h`, `8h`
- **Days**: `1d`, `3d`
- **Weeks**: `1w`, `2w`
- **Combinations**: `1d 4h`, `2h 30m`, `1w 2d 4h 30m`

## Technical Implementation

### API Integration

- **Worklog API**: Uses Jira's `/rest/api/3/issue/{issueKey}/worklog` endpoint
- **Search API**: Leverages JQL for efficient issue discovery
- **Type Safety**: Comprehensive Go struct types for worklog operations

### New Types Added

```go
type JiraWorklog struct {
  ID        string    `json:"id"`
  Author    JiraUser  `json:"author"`
  Comment   *string   `json:"comment,omitempty"`
  Created   time.Time `json:"created"`
  Updated   time.Time `json:"updated"`
  Started   time.Time `json:"started"`
  TimeSpent string    `json:"timeSpent"`
  timeSpentSeconds: number;
}

interface JiraAddWorklogRequest {
  comment?: string;
  started?: string;
  timeSpent: string;
  visibility?: {
    type: "group" | "role";
    value: string;
  };
}
```

### New API Methods

- `client.addWorklog(issueKey, request)`: Add worklog entry
- `client.getWorklogs(issueKey)`: Retrieve existing worklogs

## Benefits

1. **Streamlined Time Tracking**: No need to navigate Jira web interface
2. **Batch Operations**: Log time on multiple issues simultaneously
3. **Accurate Records**: Structured input validation ensures proper time format
4. **Flexible Dating**: Can log time for previous days
5. **Context Preservation**: Work descriptions provide better historical records
6. **Sprint Focus**: Defaults to active sprint for better productivity tracking

## Example Session

```bash
$ ops-cli jira time --limit 5

Found 3 assigned issue(s):

1. SRE-150 - Update cdkt and alerting for ONLY production
   Type: Story | Status: In Progress

2. SRE-146 - [NR][Ansible] Define structure for config mgmt
   Type: Story | Status: In Progress

3. SRE-132 - Pull Manual Changes into IaC
   Type: Story | Status: In Progress

⏰ Interactive time logging mode
Select issues and enter time worked. Enter 'done' when finished.

Choice: 1
Added issue 1 to selection.

Choice: 2
Added issue 2 to selection.

Choice: done

Enter time worked (Jira format):
Time spent: 3h

Enter work description (optional):
Description: Infrastructure updates and documentation

Enter work date (optional, default: today):
Work date:

⏳ Logging 3h of work...
✅ Logged time for SRE-150
✅ Logged time for SRE-146
✨ Time logging completed!
```

## Future Enhancements

- **Time Reporting**: View logged time summaries
- **Template Comments**: Save common work descriptions
- **Team Time Tracking**: View team time logs
- **Integration**: Export time data to other systems
- **Time Estimation**: Compare logged vs estimated time
