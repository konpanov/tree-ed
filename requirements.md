# Requirements for tree-runner



## 1. Content display
Content of a file must be displayed to the user.

### 1.1 Windows displayes part of the file that fits in it

### 1.2 Content of the file spearated by a new line sequence is displayed on a separate line

### 1.3 New line sequances are not displayed

### 1.4 Current cursor positoin must be displayed and meaningfully discribe where a change would be made


## 2. Character based navigation
User is able to navigate the file per character and per line

### 2.1 User can move cursor one character to the right (in direction of the end of a line)
#### 2.1.1 Going to the end of the line stops the cursor (cursor does not go to the next line)

### 2.2 User can move cursor one character to the left (in direction of the beginning of a line)
#### 2.2.1 Going to the start of the line stop the cursor (cursor does not go to the previous line)

### 2.3 User can move cursor one line up
#### 2.3.1 Moving up from the first line is not possible

### 2.4 Use can move cursor one line down
#### 2.4.1 Moving down from the last line is not possible

### 2.5 Saving horizontal position (column) when moving vertically (chaning row)
When moving up (2.3) or down (2.4) cursor stays on the column the movement started from
#### 2.5.1 When chaning row cursor stays on a column movement started from (origin column)
#### 2.5.2 If new row line width is smaller than origin column, cursor stays on the last column
#### 2.5.3 If vertical movements are chained, origin column is kept from the first movement
