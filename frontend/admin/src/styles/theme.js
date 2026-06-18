import { createGlobalStyle } from 'styled-components';

export const theme = {
  colors: {
    primary: '#5D3FD3', // Rich purple
    secondary: '#3F8ED3', // Bright blue
    accent: '#FF6B6B', // Coral
    success: '#4CAF50', // Green
    warning: '#FFC107', // Yellow
    error: '#F44336', // Red
    dark: '#1C1C28', // Dark blue-gray
    light: '#F8F9FA', // Light gray
    white: '#FFFFFF',
    black: '#000000',
    gray: '#6C757D',
    lightGray: '#E9ECEF',
    darkGray: '#343A40',
    gradientStart: '#5D3FD3',
    gradientEnd: '#3F8ED3'
  },
  fonts: {
    main: "'Poppins', sans-serif",
    heading: "'Montserrat', sans-serif"
  },
  fontSizes: {
    xs: '0.75rem',
    sm: '0.875rem',
    md: '1rem',
    lg: '1.125rem',
    xl: '1.25rem',
    xxl: '1.5rem',
    xxxl: '2rem',
    big: '2.5rem',
    huge: '3rem'
  },
  spacing: {
    xs: '0.25rem',
    sm: '0.5rem',
    md: '1rem',
    lg: '1.5rem',
    xl: '2rem',
    xxl: '3rem'
  },
  breakpoints: {
    xs: '0px',
    sm: '576px',
    md: '768px',
    lg: '992px',
    xl: '1200px',
    xxl: '1400px'
  },
  borderRadius: {
    sm: '0.25rem',
    md: '0.5rem',
    lg: '1rem',
    pill: '50rem',
    circle: '50%'
  },
  shadows: {
    sm: '0 1px 3px rgba(0,0,0,0.12), 0 1px 2px rgba(0,0,0,0.24)',
    md: '0 4px 6px rgba(0,0,0,0.1)',
    lg: '0 10px 15px rgba(0,0,0,0.1)',
    xl: '0 20px 25px rgba(0,0,0,0.1)',
    inner: 'inset 0 2px 4px rgba(0,0,0,0.06)'
  },
  transitions: {
    short: 'all 0.2s ease-in-out',
    medium: 'all 0.3s ease-in-out',
    long: 'all 0.5s ease-in-out'
  }
};

export const GlobalStyle = createGlobalStyle`
  /* Include the following in your public/index.html:
     <link href="https://fonts.googleapis.com/css2?family=Montserrat:wght@300;400;500;600;700&family=Poppins:wght@300;400;500;600;700&display=swap" rel="stylesheet"> */

  * {
    box-sizing: border-box;
    margin: 0;
    padding: 0;
  }

  html {
    font-size: 16px;
    scroll-behavior: smooth;
  }

  body {
    font-family: ${theme.fonts.main};
    background-color: ${theme.colors.light};
    color: ${theme.colors.dark};
    line-height: 1.5;
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
  }

  h1, h2, h3, h4, h5, h6 {
    font-family: ${theme.fonts.heading};
    margin-bottom: ${theme.spacing.md};
    font-weight: 600;
    line-height: 1.2;
  }

  p {
    margin-bottom: ${theme.spacing.md};
  }

  a {
    color: ${theme.colors.primary};
    text-decoration: none;
    transition: ${theme.transitions.short};

    &:hover {
      color: ${theme.colors.secondary};
    }
  }

  button {
    cursor: pointer;
    font-family: ${theme.fonts.main};
  }

  img {
    max-width: 100%;
    height: auto;
  }

  .container {
    width: 100%;
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 ${theme.spacing.md};
  }

  .section {
    padding: ${theme.spacing.xxl} 0;
  }

  .text-center {
    text-align: center;
  }

  .text-left {
    text-align: left;
  }

  .text-right {
    text-align: right;
  }

  .flex {
    display: flex;
  }

  .flex-col {
    flex-direction: column;
  }

  .items-center {
    align-items: center;
  }

  .justify-center {
    justify-content: center;
  }

  .justify-between {
    justify-content: space-between;
  }

  .w-full {
    width: 100%;
  }

  .h-full {
    height: 100%;
  }

  .mt-1 {
    margin-top: ${theme.spacing.xs};
  }

  .mt-2 {
    margin-top: ${theme.spacing.sm};
  }

  .mt-3 {
    margin-top: ${theme.spacing.md};
  }

  .mt-4 {
    margin-top: ${theme.spacing.lg};
  }

  .mt-5 {
    margin-top: ${theme.spacing.xl};
  }

  .mb-1 {
    margin-bottom: ${theme.spacing.xs};
  }

  .mb-2 {
    margin-bottom: ${theme.spacing.sm};
  }

  .mb-3 {
    margin-bottom: ${theme.spacing.md};
  }

  .mb-4 {
    margin-bottom: ${theme.spacing.lg};
  }

  .mb-5 {
    margin-bottom: ${theme.spacing.xl};
  }
`;

export default theme;
