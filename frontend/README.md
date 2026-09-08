# Frontend Technical Documentation

In this document, we will provide instructions on how this project developed until now and as a suggestion for further development.

## 1. Techstack

Technologies used in this project:

- React
- Vite
- TypeScript
- Material UI
- TailwindCSS
- Docker
- Nginx
- Leaflet

## 2. Structure

Project is structured as follows:

- `.vercel/`: Configuration files for Vercel deployment.
- `public/`: Static assets such as images and icons.
- `src/`: Contains the main source code for the React application.
  - `page/`: Split into directories like `Dashboard`, `Inventory`, `Orders`, etc. for specific features or modules. Each may contain:
    - `components/`: Reusable React components specific to that module.
    - `tests/`: Unit and integration tests for that module.
    - `xxx.tsx` : Main page component for that module.
  - `api/`: Contains API service files for making HTTP requests to the backend.
    - `xxxAPI.ts`: API service file for a specific module, containing functions to interact with the backend endpoints.
    - `mapper.ts`: Contains functions to map API responses to frontend data structures.
  - `utils/`: Utility functions and helpers used across the application.
  - `types/`: TypeScript type definitions and interfaces.
  - `assets/`: Static assets like images, icons, and stylesheets.
  - `tests/`: Contains test files for unit and integration testing.
  - `App.tsx`: The main application component that sets up routing and global context providers.
  - `ColorPalette.ts`: Defines the color palette used throughout the application.
  - `index.css`: Global CSS styles for the application.
  - `main.tsx`: Entry point of the React application, where the app is rendered to the DOM.
  - `vitest.setup.ts`: Configuration and setup for Vitest, the testing framework used in the project.

## 3. Conventions & Patterns

### a. File Naming

- Use PascalCase for React components (e.g., `MyComponent.tsx`).
- Use camelCase for utility functions and hooks (e.g., `useCustomHook.ts`).
- Use kebab-case for CSS and asset files (e.g., `my-component.css`).

### b. Component Structure

- Each page component should be placed in its own directory with the same name as the component.
- Each page directory should contain its own `components` and `tests` subdirectories as needed.

### c. State Management

- Using React's built-in state management like `useState` and `useReducer`
- For global state management (if needed), consider using simple state management library like Redux.

### d. API Integration

- Use the `api/` directory to organize API service files.
- For a specific module, use a dedicated API service file (e.g., `inventoryAPI.ts`) to handle all related API calls.
- Define request and response contracts in `src/types/api.ts`, then import them into API services and consumers.
- Create one clearly named function for each endpoint (e.g., `fetchInventory`, `createItem`, `createProject`).
- Every endpoint function must type its request and response, pass an optional `AbortSignal` when appropriate, and throw a useful error when the response is not successful.

### e. Error Handling
- Use try-catch blocks.
- Log errors to the console for debugging.
