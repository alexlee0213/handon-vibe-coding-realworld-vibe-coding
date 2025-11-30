import { describe, it, expect } from 'vitest';
import { render, screen } from './test/utils';
import { MantineProvider } from '@mantine/core';
import { theme } from './lib/theme';

describe('App', () => {
  it('renders without crashing', () => {
    render(
      <MantineProvider theme={theme}>
        <div data-testid="app">Conduit App</div>
      </MantineProvider>
    );
    expect(screen.getByTestId('app')).toBeInTheDocument();
  });
});
