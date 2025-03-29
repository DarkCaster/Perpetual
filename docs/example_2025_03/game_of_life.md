# Example

Creating a [Conway's Game of Life](https://playgameoflife.com/) simulation in python. In this example we will create simple game of life implementation using task mode, introduced in `v5.0.0`.

First, create new empty directory, all generated code, and other files will be placed there. It is assumed that you perform all operations with `Perpetual` from that working directory.

## Describe the task

Create text file with description of the future game: `task.md`. In this example we will use markdown formatting. It will be uploaded to the LLM as a part of the query in a text format, so plain text or simple markdown formating (without xml tags) is recommended since it can be recognized by almost any LLM.

```markdown
# Create Conway's Game of Life program with pygame library

## Main features

- The game must use the "pygame" library
- The logic should be divided into several classes: game-state, core, UI with input, predefined templates.
- Main script name: "game.py"

## Game-state class

- Delay between simulation ticks
- Grid size
- Current grid state

## UI

- Simulation step time or speed.
- Start, pause and clear buttons.
- Buttons that load predefined patterns.
- Grid size controls with grid width and height settings, min value is 10, max value is 100, default: 80x40
- The grid setting should also change the window aspect ratio accordingly so that the cells always remain square.
- Main window must be resizable and maintain aspect ratio defined by grid size.
- Place all buttons and controls at the top part of main window, grid should be below it.
- User should be able to enable and disable any grid cell by a click of a mouse.

## Predefined patterns

Please, implement following predefined patterns in a separate classes:

- Block
- Blinker
- Glider
```

## Prepare your project

### Create simple launch script

This is just for convenience, you can skip this step

<details>
<summary>run.sh</summary>

```sh
#!/bin/bash
set -e
script_dir="$( cd "$( dirname "$0" )" && pwd )"

if [[ ! -d "$script_dir/venv" ]]; then
  virtualenv "$script_dir/venv"
  "$script_dir/venv/bin/pip" --require-virtualenv install --upgrade -r "$script_dir/requirements.txt"
fi

"$script_dir/venv/bin/python" "$script_dir/game.py" "$@"
```

</details>

### Initialize the **Perpetual** project

```sh
Perpetual init -l python3
```

### Prepare your .env file

Prepare your `.env` file with your Anthropic and/or OpenAI credentials and place it in the `.perpetual` directory or [global configuration directory](../configuration.md). Use the `.perpetual/.env.example` file as a reference for all supported options. In this example, the Anthropic provider is used with the `claude-3-7-sonnet-latest` model for all operations.

### Generate Code

Generate code by running:

```sh
Perpetual implement -pr -t -i task.md
```

**Note:** using extended reasoning mode (`-pr` flag) is recommended for complex tasks or empty projects. It will ask LLM to generate work plan with initial task converted to the step-by-step instructions for the final implementation.

**Example Output:**

```text
[00.000] [INF] Project root directory: /mnt/data/Sources/GameOfLife
[00.000] [WRN] Not loading missing env file: /mnt/data/Sources/GameOfLife/.perpetual/.env
[00.001] [INF] Loaded env file: /home/user/.config/Perpetual/.env
[00.001] [INF] Fetching project files
[00.002] [INF] Calculating checksums for project files
[00.002] [INF] Annotating files, count: 1
[00.002] [INF] [provider:anthropic] [model:claude-3-7-sonnet-latest] [segments:3] [retries:3] [temperature:0.5] [max tokens:768] [think:disabled] [variants:1] [strategy:SHORT] [format:plain]
[00.002] [INF] 1: run.sh
[08.847] [INF] Saving annotations
[08.848] [INF] Running stage1: find project files for review
[08.848] [INF] [provider:anthropic] [model:claude-3-7-sonnet-latest] [segments:3] [retries:3] [temperature:0.2] [max tokens:512] [think:disabled] [variants:1] [strategy:SHORT] [format:plain]
[11.636] [INF] Files requested by LLM:
[11.636] [WRN] No matches found while salvaging filename: game.py
[11.636] [WRN] No matches found while salvaging filename: requirements.txt
[11.636] [INF] Not adding any source code files for review
[11.636] [INF] Running stage2: generating work plan
[11.636] [INF] [provider:anthropic] [model:claude-3-7-sonnet-latest] [segments:3] [retries:3] [temperature:0.5] [max tokens:1536] [think:disabled] [variants:1] [strategy:SHORT] [format:plain]
[22.413] [INF] Running stage3: generating list of files for processing
[22.413] [INF] [provider:anthropic] [model:claude-3-7-sonnet-latest] [segments:3] [retries:3] [temperature:0.2] [max tokens:512] [think:disabled] [variants:1] [strategy:SHORT] [format:plain]
[23.972] [INF] Files for processing selected by LLM:
[23.972] [INF] game.py (new file)
[23.972] [INF] game_state.py (new file)
[23.973] [INF] core.py (new file)
[23.973] [INF] ui.py (new file)
[23.973] [INF] patterns.py (new file)
[23.973] [INF] Running stage4: implementing code
[23.973] [INF] [provider:anthropic] [model:claude-3-7-sonnet-latest] [segments:3] [retries:3] [temperature:0.5] [max tokens:8192] [think:disabled] [variants:1] [strategy:SHORT] [format:plain]
[23.973] [INF] game.py
[32.962] [INF] game_state.py
[49.103] [INF] core.py
[58.019] [INF] ui.py
[112.000] [INF] patterns.py
[121.555] [INF] Creating new stash from generated results
[121.555] [INF] Processing stash: 2025-03-29_19-56-55
[121.556] [INF] Applying changes
[121.556] [INF] ui.py
[121.556] [INF] patterns.py
[121.556] [INF] game.py
[121.556] [INF] game_state.py
[121.556] [INF] core.py
```

## Generated Files

**Note**: Because LLM generation is not deterministic, you may need to run generation multiple times to achieve satisfactory results. This example is a relatively complex task for LLM as of time of writing (march 2025), so smaller local models may not work at all or produce very unstable results. Even `Claude 3.7 Sonnet` sometimes failing to produce fully working code from the first time for such a complex tasks involing generating so much code at once.

<details>
<summary>game.py</summary>

```python
#!/usr/bin/env python3
import pygame
import sys
from game_state import GameState
from core import GameCore
from ui import GameUI

def main():
    # Initialize pygame
    pygame.init()
    
    # Create game state with default values
    game_state = GameState(width=80, height=40, delay=100)
    
    # Create game core
    game_core = GameCore()
    
    # Create game UI
    game_ui = GameUI(game_state)
    
    # Game loop variables
    running = True
    simulation_active = False
    last_update_time = pygame.time.get_ticks()
    
    # Main game loop
    while running:
        current_time = pygame.time.get_ticks()
        
        # Handle events
        for event in pygame.event.get():
            if event.type == pygame.QUIT:
                running = False
            elif event.type == pygame.VIDEORESIZE:
                game_ui.handle_resize(event.w, event.h)
            else:
                # Let UI handle other events (buttons, grid clicks, etc.)
                ui_action = game_ui.handle_event(event)
                
                if ui_action == "start":
                    simulation_active = True
                elif ui_action == "pause":
                    simulation_active = False
                elif ui_action == "clear":
                    game_state.clear_grid()
                    simulation_active = False
                elif ui_action == "resize_grid":
                    # Grid was resized in the UI, update the window
                    game_ui.update_window_size()
        
        # Update game state if simulation is active and enough time has passed
        if simulation_active and current_time - last_update_time >= game_state.delay:
            game_state.grid = game_core.next_generation(game_state.grid)
            last_update_time = current_time
        
        # Draw everything
        game_ui.draw()
        
        # Update the display
        pygame.display.flip()
        
        # Cap the frame rate
        pygame.time.Clock().tick(60)
    
    # Clean up
    pygame.quit()
    sys.exit()

if __name__ == "__main__":
    main()
```

</details>

<details>
<summary>game_state.py</summary>

```python
#!/usr/bin/env python3
import numpy as np

class GameState:
    """
    Class that manages the state of Conway's Game of Life.
    Includes grid size, current grid state, and simulation delay.
    """
    
    def __init__(self, width=80, height=40, delay=100):
        """
        Initialize the game state with specified dimensions and delay.
        
        Args:
            width (int): Width of the grid (default: 80)
            height (int): Height of the grid (default: 40)
            delay (int): Delay between simulation steps in milliseconds (default: 100)
        """
        self._width = max(10, min(100, width))  # Constrain between 10 and 100
        self._height = max(10, min(100, height))  # Constrain between 10 and 100
        self._delay = delay
        self._grid = np.zeros((self._height, self._width), dtype=bool)
    
    @property
    def width(self):
        """Get the width of the grid."""
        return self._width
    
    @width.setter
    def width(self, value):
        """
        Set the width of the grid, constrained between 10 and 100.
        Resizes the grid while preserving existing cell states where possible.
        """
        value = max(10, min(100, value))  # Constrain between 10 and 100
        if value != self._width:
            old_grid = self._grid.copy()
            self._width = value
            new_grid = np.zeros((self._height, self._width), dtype=bool)
            
            # Copy the old grid data to the new grid where dimensions overlap
            h_overlap = min(old_grid.shape[0], new_grid.shape[0])
            w_overlap = min(old_grid.shape[1], new_grid.shape[1])
            new_grid[:h_overlap, :w_overlap] = old_grid[:h_overlap, :w_overlap]
            
            self._grid = new_grid
    
    @property
    def height(self):
        """Get the height of the grid."""
        return self._height
    
    @height.setter
    def height(self, value):
        """
        Set the height of the grid, constrained between 10 and 100.
        Resizes the grid while preserving existing cell states where possible.
        """
        value = max(10, min(100, value))  # Constrain between 10 and 100
        if value != self._height:
            old_grid = self._grid.copy()
            self._height = value
            new_grid = np.zeros((self._height, self._width), dtype=bool)
            
            # Copy the old grid data to the new grid where dimensions overlap
            h_overlap = min(old_grid.shape[0], new_grid.shape[0])
            w_overlap = min(old_grid.shape[1], new_grid.shape[1])
            new_grid[:h_overlap, :w_overlap] = old_grid[:h_overlap, :w_overlap]
            
            self._grid = new_grid
    
    @property
    def delay(self):
        """Get the delay between simulation steps."""
        return self._delay
    
    @delay.setter
    def delay(self, value):
        """Set the delay between simulation steps."""
        self._delay = max(10, value)  # Ensure delay is at least 10ms
    
    @property
    def grid(self):
        """Get the current grid state."""
        return self._grid
    
    @grid.setter
    def grid(self, new_grid):
        """Set the grid to a new state."""
        if new_grid.shape == (self._height, self._width):
            self._grid = new_grid
        else:
            raise ValueError(f"New grid dimensions {new_grid.shape} do not match expected dimensions ({self._height}, {self._width})")
    
    def toggle_cell(self, row, col):
        """
        Toggle the state of a cell at the specified position.
        
        Args:
            row (int): Row index of the cell
            col (int): Column index of the cell
        
        Returns:
            bool: The new state of the cell
        """
        if 0 <= row < self._height and 0 <= col < self._width:
            self._grid[row, col] = not self._grid[row, col]
            return self._grid[row, col]
        return False
    
    def set_cell(self, row, col, state):
        """
        Set the state of a cell at the specified position.
        
        Args:
            row (int): Row index of the cell
            col (int): Column index of the cell
            state (bool): New state for the cell
        
        Returns:
            bool: True if the cell was set, False if out of bounds
        """
        if 0 <= row < self._height and 0 <= col < self._width:
            self._grid[row, col] = state
            return True
        return False
    
    def clear_grid(self):
        """Clear the grid by setting all cells to False."""
        self._grid.fill(False)
    
    def resize_grid(self, width, height):
        """
        Resize the grid to the specified dimensions.
        
        Args:
            width (int): New width for the grid
            height (int): New height for the grid
        """
        self.width = width
        self.height = height
```

</details>

<details>
<summary>core.py</summary>

```python
#!/usr/bin/env python3
import numpy as np

class GameCore:
    """
    Core logic for Conway's Game of Life.
    Implements the rules and calculates the next generation of the grid.
    """
    
    def __init__(self):
        """Initialize the GameCore."""
        pass
    
    def next_generation(self, grid):
        """
        Calculate the next generation of the grid based on Conway's Game of Life rules:
        1. Any live cell with fewer than two live neighbors dies (underpopulation)
        2. Any live cell with two or three live neighbors lives on
        3. Any live cell with more than three live neighbors dies (overpopulation)
        4. Any dead cell with exactly three live neighbors becomes alive (reproduction)
        
        Args:
            grid (numpy.ndarray): Current grid state
            
        Returns:
            numpy.ndarray: Next generation grid state
        """
        # Create a copy of the grid to store the next generation
        next_grid = np.copy(grid)
        
        # Get grid dimensions
        height, width = grid.shape
        
        # Calculate the number of live neighbors for each cell
        for row in range(height):
            for col in range(width):
                # Count live neighbors (including wrapping around the edges)
                live_neighbors = self._count_live_neighbors(grid, row, col)
                
                # Apply Conway's Game of Life rules
                if grid[row, col]:  # Cell is alive
                    if live_neighbors < 2 or live_neighbors > 3:
                        next_grid[row, col] = False  # Cell dies
                else:  # Cell is dead
                    if live_neighbors == 3:
                        next_grid[row, col] = True  # Cell becomes alive
        
        return next_grid
    
    def _count_live_neighbors(self, grid, row, col):
        """
        Count the number of live neighbors for a cell, including wrapping around edges.
        
        Args:
            grid (numpy.ndarray): Current grid state
            row (int): Row index of the cell
            col (int): Column index of the cell
            
        Returns:
            int: Number of live neighbors
        """
        height, width = grid.shape
        count = 0
        
        # Check all 8 neighboring cells
        for i in range(-1, 2):
            for j in range(-1, 2):
                if i == 0 and j == 0:
                    continue  # Skip the cell itself
                
                # Calculate neighbor coordinates with wrapping
                neighbor_row = (row + i) % height
                neighbor_col = (col + j) % width
                
                # Count if the neighbor is alive
                if grid[neighbor_row, neighbor_col]:
                    count += 1
        
        return count
```

</details>

<details>
<summary>ui.py</summary>

```python
#!/usr/bin/env python3
import pygame
import sys
from patterns import Block, Blinker, Glider

class GameUI:
    """
    User interface for Conway's Game of Life.
    Handles rendering, user input, and UI controls.
    """
    
    def __init__(self, game_state):
        """
        Initialize the game UI.
        
        Args:
            game_state (GameState): The game state to render and control
        """
        self.game_state = game_state
        
        # UI colors
        self.BACKGROUND_COLOR = (240, 240, 240)
        self.GRID_COLOR = (200, 200, 200)
        self.CELL_COLOR = (50, 50, 50)
        self.BUTTON_COLOR = (100, 100, 200)
        self.BUTTON_HOVER_COLOR = (120, 120, 220)
        self.BUTTON_TEXT_COLOR = (255, 255, 255)
        self.CONTROL_PANEL_COLOR = (220, 220, 220)
        
        # UI dimensions
        self.CONTROL_PANEL_HEIGHT = 100
        self.CELL_SIZE = 10  # Initial cell size, will be adjusted based on window size
        self.BUTTON_HEIGHT = 30
        self.BUTTON_PADDING = 10
        
        # Calculate initial window size based on grid dimensions
        window_width = max(800, game_state.width * self.CELL_SIZE)
        window_height = self.CONTROL_PANEL_HEIGHT + game_state.height * self.CELL_SIZE
        
        # Create the window
        self.window = pygame.display.set_mode(
            (window_width, window_height),
            pygame.RESIZABLE
        )
        pygame.display.set_caption("Conway's Game of Life")
        
        # Initialize fonts
        pygame.font.init()
        self.font = pygame.font.SysFont('Arial', 16)

        # Patterns
        self.patterns = {
            "Block": Block(),
            "Blinker": Blinker(),
            "Glider": Glider()
        }

        # Create UI elements
        self._create_ui_elements()
        
        # Simulation state
        self.simulation_active = False
        

    def _create_ui_elements(self):
        """Create all UI buttons and controls."""
        # Button dimensions and positions
        button_width = 100
        slider_width = 150
        x_pos = self.BUTTON_PADDING
        y_pos = self.BUTTON_PADDING
        
        # Create buttons
        self.buttons = []
        
        # Start/Pause button
        self.start_pause_button = {
            "rect": pygame.Rect(x_pos, y_pos, button_width, self.BUTTON_HEIGHT),
            "text": "Start",
            "action": "start",
            "hover": False
        }
        self.buttons.append(self.start_pause_button)
        x_pos += button_width + self.BUTTON_PADDING
        
        # Clear button
        self.clear_button = {
            "rect": pygame.Rect(x_pos, y_pos, button_width, self.BUTTON_HEIGHT),
            "text": "Clear",
            "action": "clear",
            "hover": False
        }
        self.buttons.append(self.clear_button)
        x_pos += button_width + self.BUTTON_PADDING
        
        # Pattern buttons
        for pattern_name in self.patterns.keys():
            pattern_button = {
                "rect": pygame.Rect(x_pos, y_pos, button_width, self.BUTTON_HEIGHT),
                "text": pattern_name,
                "action": f"pattern_{pattern_name}",
                "hover": False
            }
            self.buttons.append(pattern_button)
            x_pos += button_width + self.BUTTON_PADDING
        
        # Second row for sliders and grid controls
        x_pos = self.BUTTON_PADDING
        y_pos += self.BUTTON_HEIGHT + self.BUTTON_PADDING
        
        # Speed slider
        self.speed_slider = {
            "rect": pygame.Rect(x_pos, y_pos, slider_width, self.BUTTON_HEIGHT),
            "label": "Speed:",
            "value": 100,  # Initial delay in ms (lower = faster)
            "min": 10,
            "max": 500,
            "dragging": False
        }
        x_pos += slider_width + self.BUTTON_PADDING * 2
        
        # Grid width slider
        self.width_slider = {
            "rect": pygame.Rect(x_pos, y_pos, slider_width, self.BUTTON_HEIGHT),
            "label": "Width:",
            "value": self.game_state.width,
            "min": 10,
            "max": 100,
            "dragging": False
        }
        x_pos += slider_width + self.BUTTON_PADDING * 2
        
        # Grid height slider
        self.height_slider = {
            "rect": pygame.Rect(x_pos, y_pos, slider_width, self.BUTTON_HEIGHT),
            "label": "Height:",
            "value": self.game_state.height,
            "min": 10,
            "max": 100,
            "dragging": False
        }
    
    def handle_event(self, event):
        """
        Handle pygame events.
        
        Args:
            event (pygame.event.Event): The event to handle
            
        Returns:
            str or None: Action to perform, or None if no action
        """
        if event.type == pygame.MOUSEBUTTONDOWN:
            # Check if a button was clicked
            for button in self.buttons:
                if button["rect"].collidepoint(event.pos):
                    if button["action"] == "start" and not self.simulation_active:
                        self.simulation_active = True
                        self.start_pause_button["text"] = "Pause"
                        self.start_pause_button["action"] = "pause"
                        return "start"
                    elif button["action"] == "pause" and self.simulation_active:
                        self.simulation_active = False
                        self.start_pause_button["text"] = "Start"
                        self.start_pause_button["action"] = "start"
                        return "pause"
                    elif button["action"] == "clear":
                        self.simulation_active = False
                        self.start_pause_button["text"] = "Start"
                        self.start_pause_button["action"] = "start"
                        return "clear"
                    elif button["action"].startswith("pattern_"):
                        pattern_name = button["action"].split("_")[1]
                        self._apply_pattern(pattern_name)
                        return f"pattern_{pattern_name}"
            
            # Check if a slider was clicked
            if self.speed_slider["rect"].collidepoint(event.pos):
                self.speed_slider["dragging"] = True
                self._update_slider_value(self.speed_slider, event.pos[0])
                self.game_state.delay = self.speed_slider["value"]
            
            if self.width_slider["rect"].collidepoint(event.pos):
                self.width_slider["dragging"] = True
                self._update_slider_value(self.width_slider, event.pos[0])
            
            if self.height_slider["rect"].collidepoint(event.pos):
                self.height_slider["dragging"] = True
                self._update_slider_value(self.height_slider, event.pos[0])
            
            # Check if the grid was clicked
            if self._is_click_on_grid(event.pos):
                grid_pos = self._screen_to_grid_pos(event.pos)
                self.game_state.toggle_cell(grid_pos[1], grid_pos[0])
                return "grid_click"
        
        elif event.type == pygame.MOUSEBUTTONUP:
            # Stop dragging sliders
            if self.speed_slider["dragging"]:
                self.speed_slider["dragging"] = False
            
            if self.width_slider["dragging"]:
                self.width_slider["dragging"] = False
                if self.width_slider["value"] != self.game_state.width:
                    self.game_state.width = self.width_slider["value"]
                    return "resize_grid"
            
            if self.height_slider["dragging"]:
                self.height_slider["dragging"] = False
                if self.height_slider["value"] != self.game_state.height:
                    self.game_state.height = self.height_slider["value"]
                    return "resize_grid"
        
        elif event.type == pygame.MOUSEMOTION:
            # Update button hover states
            for button in self.buttons:
                button["hover"] = button["rect"].collidepoint(event.pos)
            
            # Update slider values if dragging
            if self.speed_slider["dragging"]:
                self._update_slider_value(self.speed_slider, event.pos[0])
                self.game_state.delay = self.speed_slider["value"]
            
            if self.width_slider["dragging"]:
                self._update_slider_value(self.width_slider, event.pos[0])
            
            if self.height_slider["dragging"]:
                self._update_slider_value(self.height_slider, event.pos[0])
        
        return None
    
    def _update_slider_value(self, slider, x_pos):
        """
        Update a slider's value based on mouse position.
        
        Args:
            slider (dict): The slider to update
            x_pos (int): Mouse x position
        """
        slider_left = slider["rect"].left
        slider_width = slider["rect"].width
        value_range = slider["max"] - slider["min"]
        
        # Calculate new value based on mouse position
        relative_pos = max(0, min(slider_width, x_pos - slider_left))
        slider["value"] = slider["min"] + int((relative_pos / slider_width) * value_range)
    
    def _apply_pattern(self, pattern_name):
        """
        Apply a predefined pattern to the grid.
        
        Args:
            pattern_name (str): Name of the pattern to apply
        """
        if pattern_name in self.patterns:
            # Calculate center position
            center_row = self.game_state.height // 2
            center_col = self.game_state.width // 2
            
            # Apply the pattern
            self.patterns[pattern_name].apply(self.game_state, center_col, center_row)
    
    def _is_click_on_grid(self, pos):
        """
        Check if a click position is on the grid.
        
        Args:
            pos (tuple): (x, y) position of the click
            
        Returns:
            bool: True if the click is on the grid, False otherwise
        """
        x, y = pos
        grid_top = self.CONTROL_PANEL_HEIGHT
        
        return (y >= grid_top and 
                x >= 0 and 
                x < self.window.get_width() and 
                y < self.window.get_height())
    
    def _screen_to_grid_pos(self, pos):
        """
        Convert screen coordinates to grid coordinates.
        
        Args:
            pos (tuple): (x, y) position on the screen
            
        Returns:
            tuple: (col, row) position on the grid
        """
        x, y = pos
        grid_top = self.CONTROL_PANEL_HEIGHT
        
        col = int(x / self.CELL_SIZE)
        row = int((y - grid_top) / self.CELL_SIZE)
        
        # Ensure the coordinates are within the grid bounds
        col = max(0, min(col, self.game_state.width - 1))
        row = max(0, min(row, self.game_state.height - 1))
        
        return (col, row)
    
    def handle_resize(self, width, height):
        """
        Handle window resize event.
        
        Args:
            width (int): New window width
            height (int): New window height
        """
        # Ensure minimum window size
        width = max(400, width)
        height = max(300, height)
        
        # Resize the window
        self.window = pygame.display.set_mode((width, height), pygame.RESIZABLE)
        
        # Recalculate cell size
        self._recalculate_cell_size()
        
        # Update UI element positions
        self._create_ui_elements()
    
    def update_window_size(self):
        """Update window size based on grid dimensions."""
        # Calculate new window size
        grid_width = self.game_state.width * self.CELL_SIZE
        grid_height = self.game_state.height * self.CELL_SIZE
        
        window_width = max(800, grid_width)
        window_height = self.CONTROL_PANEL_HEIGHT + grid_height
        
        # Resize the window
        self.window = pygame.display.set_mode((window_width, window_height), pygame.RESIZABLE)
        
        # Update UI element positions
        self._create_ui_elements()
    
    def _recalculate_cell_size(self):
        """Recalculate cell size based on window dimensions and grid size."""
        window_width = self.window.get_width()
        window_height = self.window.get_height() - self.CONTROL_PANEL_HEIGHT
        
        # Calculate cell size to fit the grid in the window
        width_cell_size = window_width / self.game_state.width
        height_cell_size = window_height / self.game_state.height
        
        # Use the smaller of the two to ensure cells are square and fit in the window
        self.CELL_SIZE = min(width_cell_size, height_cell_size)
    
    def draw(self):
        """Draw the game UI and grid."""
        # Clear the window
        self.window.fill(self.BACKGROUND_COLOR)
        
        # Draw control panel background
        pygame.draw.rect(
            self.window,
            self.CONTROL_PANEL_COLOR,
            (0, 0, self.window.get_width(), self.CONTROL_PANEL_HEIGHT)
        )
        
        # Draw buttons
        for button in self.buttons:
            color = self.BUTTON_HOVER_COLOR if button["hover"] else self.BUTTON_COLOR
            pygame.draw.rect(self.window, color, button["rect"])
            
            text = self.font.render(button["text"], True, self.BUTTON_TEXT_COLOR)
            text_rect = text.get_rect(center=button["rect"].center)
            self.window.blit(text, text_rect)
        
        # Draw sliders
        self._draw_slider(self.speed_slider)
        self._draw_slider(self.width_slider)
        self._draw_slider(self.height_slider)
        
        # Draw grid
        self._draw_grid()
    
    def _draw_slider(self, slider):
        """
        Draw a slider control.
        
        Args:
            slider (dict): The slider to draw
        """
        # Draw slider background
        pygame.draw.rect(self.window, (180, 180, 180), slider["rect"])
        
        # Draw slider label
        label_text = self.font.render(f"{slider['label']} {slider['value']}", True, (0, 0, 0))
        label_rect = label_text.get_rect(midleft=(slider["rect"].left, slider["rect"].centery))
        self.window.blit(label_text, label_rect)
        
        # Draw slider handle
        handle_pos = slider["rect"].left + ((slider["value"] - slider["min"]) / 
                                           (slider["max"] - slider["min"])) * slider["rect"].width
        handle_rect = pygame.Rect(handle_pos - 5, slider["rect"].top, 10, slider["rect"].height)
        pygame.draw.rect(self.window, self.BUTTON_COLOR, handle_rect)
    
    def _draw_grid(self):
        """Draw the game grid and cells."""
        grid_top = self.CONTROL_PANEL_HEIGHT
        grid_width = self.game_state.width * self.CELL_SIZE
        grid_height = self.game_state.height * self.CELL_SIZE
        
        # Draw grid background
        pygame.draw.rect(
            self.window,
            self.BACKGROUND_COLOR,
            (0, grid_top, grid_width, grid_height)
        )
        
        # Draw grid lines
        for x in range(0, self.game_state.width + 1):
            pygame.draw.line(
                self.window,
                self.GRID_COLOR,
                (x * self.CELL_SIZE, grid_top),
                (x * self.CELL_SIZE, grid_top + grid_height)
            )
        
        for y in range(0, self.game_state.height + 1):
            pygame.draw.line(
                self.window,
                self.GRID_COLOR,
                (0, grid_top + y * self.CELL_SIZE),
                (grid_width, grid_top + y * self.CELL_SIZE)
            )
        
        # Draw live cells
        for row in range(self.game_state.height):
            for col in range(self.game_state.width):
                if self.game_state.grid[row, col]:
                    pygame.draw.rect(
                        self.window,
                        self.CELL_COLOR,
                        (col * self.CELL_SIZE + 1,
                         grid_top + row * self.CELL_SIZE + 1,
                         self.CELL_SIZE - 1,
                         self.CELL_SIZE - 1)
                    )
```

</details>

<details>
<summary>patterns.py</summary>

```python
#!/usr/bin/env python3

class Pattern:
    """
    Base class for predefined patterns in Conway's Game of Life.
    """
    
    def __init__(self):
        """Initialize the pattern."""
        self.cells = []  # List of (x, y) relative coordinates for live cells
    
    def apply(self, game_state, x, y):
        """
        Apply the pattern to the game grid at the specified position.
        
        Args:
            game_state: The game state to modify
            x (int): X-coordinate (column) for the pattern center
            y (int): Y-coordinate (row) for the pattern center
        """
        # Clear the grid first
        game_state.clear_grid()
        
        # Set cells according to the pattern
        for dx, dy in self.cells:
            # Calculate absolute position
            abs_x = x + dx
            abs_y = y + dy
            
            # Apply with wrapping around grid edges
            wrapped_x = abs_x % game_state.width
            wrapped_y = abs_y % game_state.height
            
            # Set the cell to alive
            game_state.set_cell(wrapped_y, wrapped_x, True)


class Block(Pattern):
    """
    Block pattern: A 2x2 still life pattern.
    
    ##
    ##
    """
    
    def __init__(self):
        super().__init__()
        self.cells = [
            (0, 0), (1, 0),
            (0, 1), (1, 1)
        ]


class Blinker(Pattern):
    """
    Blinker pattern: A period 2 oscillator.
    
    Initial state:
    ###
    
    Next state:
    #
    #
    #
    """
    
    def __init__(self):
        super().__init__()
        self.cells = [
            (-1, 0), (0, 0), (1, 0)
        ]


class Glider(Pattern):
    """
    Glider pattern: A pattern that moves diagonally across the grid.
    
    .#.
    ..#
    ###
    """
    
    def __init__(self):
        super().__init__()
        self.cells = [
            (0, -1),
            (1, 0),
            (-1, 1), (0, 1), (1, 1)
        ]
```

</details>

