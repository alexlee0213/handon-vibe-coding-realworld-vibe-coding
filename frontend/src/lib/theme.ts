import { createTheme, type MantineColorsTuple } from '@mantine/core';

// RealWorld brand color (green)
const brandGreen: MantineColorsTuple = [
  '#e6f9ed',
  '#d0f1dc',
  '#a3e4bb',
  '#72d697',
  '#4bcb79',
  '#35c466',
  '#28c05c',
  '#1ca94c',
  '#109642',
  '#008236',
];

export const theme = createTheme({
  primaryColor: 'brand',
  colors: {
    brand: brandGreen,
  },
  fontFamily: 'source sans pro, sans-serif',
  headings: {
    fontFamily: 'titillium web, sans-serif',
  },
  defaultRadius: 'sm',
});
