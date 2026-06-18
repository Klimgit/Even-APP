import React from 'react';
import styled from 'styled-components';
import { Link as RouterLink } from 'react-router-dom';

// Button Component
export const Button = styled.button`
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.5rem 1.5rem;
  font-size: ${({ theme }) => theme.fontSizes.md};
  font-weight: 500;
  border-radius: ${({ theme }) => theme.borderRadius.md};
  transition: ${({ theme }) => theme.transitions.medium};
  cursor: pointer;
  border: none;
  text-align: center;
  min-width: 120px;
  
  ${({ variant, theme }) => {
    switch (variant) {
      case 'primary':
        return `
          background-color: ${theme.colors.primary};
          color: ${theme.colors.white};
          &:hover {
            background-color: ${theme.colors.secondary};
          }
        `;
      case 'secondary':
        return `
          background-color: ${theme.colors.secondary};
          color: ${theme.colors.white};
          &:hover {
            background-color: ${theme.colors.primary};
          }
        `;
      case 'outline':
        return `
          background-color: transparent;
          color: ${theme.colors.primary};
          border: 2px solid ${theme.colors.primary};
          &:hover {
            background-color: ${theme.colors.primary};
            color: ${theme.colors.white};
          }
        `;
      case 'text':
        return `
          background-color: transparent;
          color: ${theme.colors.primary};
          padding: 0.5rem 1rem;
          min-width: auto;
          &:hover {
            color: ${theme.colors.secondary};
            text-decoration: underline;
          }
        `;
      default:
        return `
          background-color: ${theme.colors.primary};
          color: ${theme.colors.white};
          &:hover {
            background-color: ${theme.colors.secondary};
          }
        `;
    }
  }}
  
  ${({ size, theme }) => {
    switch (size) {
      case 'sm':
        return `
          padding: 0.25rem 1rem;
          font-size: ${theme.fontSizes.sm};
          min-width: 100px;
        `;
      case 'lg':
        return `
          padding: 0.75rem 2rem;
          font-size: ${theme.fontSizes.lg};
          min-width: 140px;
        `;
      default:
        return '';
    }
  }}
  
  ${({ fullWidth }) => fullWidth && `
    width: 100%;
  `}
  
  &:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
`;

// Input Component
export const Input = styled.input`
  width: 100%;
  padding: 0.75rem 1rem;
  font-size: ${({ theme }) => theme.fontSizes.md};
  border: 1px solid ${({ theme }) => theme.colors.lightGray};
  border-radius: ${({ theme }) => theme.borderRadius.md};
  background-color: ${({ theme }) => theme.colors.white};
  transition: ${({ theme }) => theme.transitions.short};
  color: ${({ theme }) => theme.colors.dark};
  
  &:focus {
    outline: none;
    border-color: ${({ theme }) => theme.colors.primary};
    box-shadow: 0 0 0 2px rgba(93, 63, 211, 0.2);
  }
  
  &::placeholder {
    color: ${({ theme }) => theme.colors.gray};
  }
  
  &:disabled {
    background-color: ${({ theme }) => theme.colors.lightGray};
    cursor: not-allowed;
  }
  
  ${({ error, theme }) => error && `
    border-color: ${theme.colors.error};
    
    &:focus {
      box-shadow: 0 0 0 2px rgba(244, 67, 54, 0.2);
    }
  `}
`;

// Form Group
export const FormGroup = styled.div`
  margin-bottom: 1.5rem;
  width: 100%;
`;

// Form Label
export const Label = styled.label`
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
  color: ${({ theme }) => theme.colors.darkGray};
`;

// Error Message
export const ErrorMessage = styled.p`
  color: ${({ theme }) => theme.colors.error};
  font-size: ${({ theme }) => theme.fontSizes.sm};
  margin-top: 0.25rem;
  margin-bottom: 0;
`;

// Card Component
export const Card = styled.div`
  background-color: ${({ theme }) => theme.colors.white};
  border-radius: ${({ theme }) => theme.borderRadius.lg};
  box-shadow: ${({ theme }) => theme.shadows.md};
  padding: 2rem;
  width: 100%;
  transition: ${({ theme }) => theme.transitions.medium};
  
  &:hover {
    box-shadow: ${({ theme }) => theme.shadows.lg};
  }
`;

// Container
export const Container = styled.div`
  width: 100%;
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 1rem;
  
  @media (min-width: ${({ theme }) => theme.breakpoints.md}) {
    padding: 0 2rem;
  }
`;

// Flex Container
export const Flex = styled.div`
  display: flex;
  flex-direction: ${({ direction }) => direction || 'row'};
  align-items: ${({ align }) => align || 'stretch'};
  justify-content: ${({ justify }) => justify || 'flex-start'};
  flex-wrap: ${({ wrap }) => wrap || 'nowrap'};
  gap: ${({ gap, theme }) => gap ? theme.spacing[gap] || gap : '0'};
`;

// Link
export const Link = styled(RouterLink)`
  color: ${({ theme }) => theme.colors.primary};
  text-decoration: none;
  transition: ${({ theme }) => theme.transitions.short};
  
  &:hover {
    color: ${({ theme }) => theme.colors.secondary};
    text-decoration: underline;
  }
`;

// Title
export const Title = styled.h1`
  font-family: ${({ theme }) => theme.fonts.heading};
  font-weight: 700;
  margin-bottom: 1rem;
  color: ${({ theme }) => theme.colors.dark};
  font-size: ${({ size, theme }) => {
    switch (size) {
      case 'sm':
        return theme.fontSizes.xl;
      case 'md':
        return theme.fontSizes.xxl;
      case 'lg':
        return theme.fontSizes.xxxl;
      case 'xl':
        return theme.fontSizes.big;
      default:
        return theme.fontSizes.xxl;
    }
  }};
  
  ${({ center }) => center && 'text-align: center;'}
  ${({ color, theme }) => color && `color: ${theme.colors[color] || color};`}
`;

// Subtitle
export const Subtitle = styled.h2`
  font-family: ${({ theme }) => theme.fonts.heading};
  font-weight: 600;
  margin-bottom: 1rem;
  color: ${({ theme }) => theme.colors.darkGray};
  font-size: ${({ size, theme }) => {
    switch (size) {
      case 'sm':
        return theme.fontSizes.lg;
      case 'md':
        return theme.fontSizes.xl;
      case 'lg':
        return theme.fontSizes.xxl;
      default:
        return theme.fontSizes.xl;
    }
  }};
  
  ${({ center }) => center && 'text-align: center;'}
  ${({ color, theme }) => color && `color: ${theme.colors[color] || color};`}
`;

// Text
export const Text = styled.p`
  font-family: ${({ theme }) => theme.fonts.main};
  margin-bottom: 1rem;
  color: ${({ theme }) => theme.colors.dark};
  font-size: ${({ size, theme }) => {
    switch (size) {
      case 'sm':
        return theme.fontSizes.sm;
      case 'md':
        return theme.fontSizes.md;
      case 'lg':
        return theme.fontSizes.lg;
      default:
        return theme.fontSizes.md;
    }
  }};
  
  ${({ center }) => center && 'text-align: center;'}
  ${({ color, theme }) => color && `color: ${theme.colors[color] || color};`}
  ${({ bold }) => bold && 'font-weight: 600;'}
`;

// Divider
export const Divider = styled.hr`
  border: 0;
  height: 1px;
  background-color: ${({ theme }) => theme.colors.lightGray};
  margin: 1.5rem 0;
`;

// Social Button
export const SocialButton = styled.button`
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  padding: 0.75rem;
  border-radius: ${({ theme }) => theme.borderRadius.md};
  font-size: ${({ theme }) => theme.fontSizes.md};
  font-weight: 500;
  transition: ${({ theme }) => theme.transitions.short};
  border: 1px solid ${({ theme }) => theme.colors.lightGray};
  background-color: ${({ theme }) => theme.colors.white};
  color: ${({ theme }) => theme.colors.darkGray};
  cursor: pointer;
  
  &:hover {
    background-color: ${({ theme }) => theme.colors.lightGray};
  }
  
  svg {
    margin-right: 0.5rem;
  }
`;

// Logo
export const Logo = styled.div`
  font-family: ${({ theme }) => theme.fonts.heading};
  font-weight: 700;
  font-size: ${({ theme }) => theme.fontSizes.xxl};
  color: ${({ theme }) => theme.colors.primary};
  
  span {
    color: ${({ theme }) => theme.colors.secondary};
  }
`;

export default {
  Button,
  Input,
  FormGroup,
  Label,
  ErrorMessage,
  Card,
  Container,
  Flex,
  Link,
  Title,
  Subtitle,
  Text,
  Divider,
  SocialButton,
  Logo
};
