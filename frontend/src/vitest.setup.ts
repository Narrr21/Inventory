import '@testing-library/jest-dom';
import { beforeAll } from 'vitest';

const originalError = console.error;

beforeAll(() => {
  console.error = (...args: unknown[]) => {
    const message = typeof args[0] === 'string' ? args[0] : '';
    if (message.includes('not configured to support act')) {
      return; // known unresolved RTL+Vitest+React issue, tests still pass
    }
    originalError(...args);
  };
});