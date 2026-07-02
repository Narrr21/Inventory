import { render, screen } from "@testing-library/react";
import { expect, test } from "vitest";
import App from "../App";

test("renders get started heading", () => {
  render(<App />);
  const headingElement = screen.getByText(/Get started/i);
  expect(headingElement).toBeInTheDocument();
});
