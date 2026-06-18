import React, { useState } from 'react';
import styled from 'styled-components';
import { Link, useNavigate } from 'react-router-dom';
import { FaBars, FaTimes, FaUserCircle } from 'react-icons/fa';
import { useAuth } from '../context/AuthContext';
import { logoutUser } from '../firebase/authService';
import { Container, Button } from '../components/ui';

const HeaderWrapper = styled.header`
  background-color: ${({ theme }) => theme.colors.white};
  box-shadow: ${({ theme }) => theme.shadows.sm};
  position: sticky;
  top: 0;
  z-index: 1000;
`;

const HeaderContainer = styled(Container)`
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 1rem;
  padding-bottom: 1rem;
`;

const Logo = styled(Link)`
  font-family: ${({ theme }) => theme.fonts.heading};
  font-weight: 700;
  font-size: ${({ theme }) => theme.fontSizes.xl};
  color: ${({ theme }) => theme.colors.primary};
  text-decoration: none;
  
  span {
    color: ${({ theme }) => theme.colors.secondary};
  }
`;

const NavLinks = styled.nav`
  display: none;
  
  @media (min-width: ${({ theme }) => theme.breakpoints.md}) {
    display: flex;
    align-items: center;
    gap: 1.5rem;
  }
`;

const NavLink = styled(Link)`
  color: ${({ theme }) => theme.colors.darkGray};
  font-weight: 500;
  text-decoration: none;
  transition: ${({ theme }) => theme.transitions.short};
  
  &:hover, &.active {
    color: ${({ theme }) => theme.colors.primary};
  }
`;

const MobileMenuButton = styled.button`
  display: flex;
  background: none;
  border: none;
  color: ${({ theme }) => theme.colors.darkGray};
  font-size: 1.5rem;
  cursor: pointer;
  
  @media (min-width: ${({ theme }) => theme.breakpoints.md}) {
    display: none;
  }
`;

const MobileMenu = styled.div`
  position: fixed;
  top: 0;
  right: 0;
  bottom: 0;
  width: 80%;
  max-width: 300px;
  background-color: ${({ theme }) => theme.colors.white};
  box-shadow: ${({ theme }) => theme.shadows.lg};
  z-index: 1100;
  transform: translateX(${({ isOpen }) => (isOpen ? '0' : '100%')});
  transition: transform 0.3s ease-in-out;
  display: flex;
  flex-direction: column;
  padding: 2rem;
`;

const MobileNavLinks = styled.nav`
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  margin-top: 2rem;
`;

const MobileNavLink = styled(Link)`
  color: ${({ theme }) => theme.colors.darkGray};
  font-weight: 500;
  text-decoration: none;
  font-size: ${({ theme }) => theme.fontSizes.lg};
  transition: ${({ theme }) => theme.transitions.short};
  
  &:hover, &.active {
    color: ${({ theme }) => theme.colors.primary};
  }
`;

const CloseButton = styled.button`
  position: absolute;
  top: 1rem;
  right: 1rem;
  background: none;
  border: none;
  color: ${({ theme }) => theme.colors.darkGray};
  font-size: 1.5rem;
  cursor: pointer;
`;

const Overlay = styled.div`
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  z-index: 1050;
  opacity: ${({ isOpen }) => (isOpen ? '1' : '0')};
  visibility: ${({ isOpen }) => (isOpen ? 'visible' : 'hidden')};
  transition: opacity 0.3s ease-in-out, visibility 0.3s ease-in-out;
`;

const UserMenu = styled.div`
  position: relative;
  display: inline-block;
`;

const UserButton = styled.button`
  background: none;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: ${({ theme }) => theme.colors.darkGray};
  
  svg {
    font-size: 1.5rem;
  }
`;

const UserDropdown = styled.div`
  position: absolute;
  top: 100%;
  right: 0;
  background-color: ${({ theme }) => theme.colors.white};
  border-radius: ${({ theme }) => theme.borderRadius.md};
  box-shadow: ${({ theme }) => theme.shadows.md};
  min-width: 200px;
  z-index: 1000;
  display: ${({ isOpen }) => (isOpen ? 'block' : 'none')};
  margin-top: 0.5rem;
  overflow: hidden;
`;

const DropdownItem = styled.button`
  display: block;
  width: 100%;
  padding: 0.75rem 1rem;
  text-align: left;
  background: none;
  border: none;
  cursor: pointer;
  color: ${({ theme }) => theme.colors.darkGray};
  transition: ${({ theme }) => theme.transitions.short};
  
  &:hover {
    background-color: ${({ theme }) => theme.colors.lightGray};
  }
`;

const UserInfo = styled.div`
  padding: 1rem;
  border-bottom: 1px solid ${({ theme }) => theme.colors.lightGray};
  
  p {
    margin: 0;
    font-weight: 500;
  }
  
  span {
    font-size: ${({ theme }) => theme.fontSizes.sm};
    color: ${({ theme }) => theme.colors.gray};
  }
`;

const Header = () => {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const [userMenuOpen, setUserMenuOpen] = useState(false);
  const { isAuthenticated, currentUser, userData } = useAuth();
  const navigate = useNavigate();
  
  const toggleMobileMenu = () => {
    setMobileMenuOpen(!mobileMenuOpen);
  };
  
  const toggleUserMenu = () => {
    setUserMenuOpen(!userMenuOpen);
  };
  
  const handleLogout = async () => {
    try {
      await logoutUser();
      setUserMenuOpen(false);
      navigate('/signin');
    } catch (error) {
      console.error('Error logging out:', error);
    }
  };
  
  const handleClickOutside = () => {
    setUserMenuOpen(false);
  };
  
  return (
    <HeaderWrapper>
      <HeaderContainer>
        <Logo to="/">
          Learn<span>Hub</span>
        </Logo>
        
        <NavLinks>
          <NavLink to="/courses">Courses</NavLink>
          <NavLink to="/instructors">Instructors</NavLink>
          <NavLink to="/about">About</NavLink>
          <NavLink to="/admin">Admin Panel</NavLink>
          
          {isAuthenticated ? (
            <UserMenu>
              <UserButton onClick={toggleUserMenu}>
                <FaUserCircle />
                <span>{currentUser?.displayName?.split(' ')[0] || 'User'}</span>
              </UserButton>
              
              <UserDropdown isOpen={userMenuOpen}>
                <UserInfo>
                  <p>{userData?.fullName || currentUser?.displayName}</p>
                  <span>{userData?.email || currentUser?.email}</span>
                </UserInfo>
                <DropdownItem onClick={() => {
                  setUserMenuOpen(false);
                  navigate('/dashboard');
                }}>
                  Dashboard
                </DropdownItem>
                <DropdownItem onClick={() => {
                  setUserMenuOpen(false);
                  navigate('/profile');
                }}>
                  Profile
                </DropdownItem>
                <DropdownItem onClick={handleLogout}>
                  Log Out
                </DropdownItem>
              </UserDropdown>
            </UserMenu>
          ) : (
            <>
              <NavLink to="/signin">Sign In</NavLink>
              <Button 
                as={Link} 
                to="/signup" 
                variant="primary"
                size="sm"
              >
                Sign Up
              </Button>
            </>
          )}
        </NavLinks>
        
        <MobileMenuButton onClick={toggleMobileMenu}>
          <FaBars />
        </MobileMenuButton>
        
        <Overlay isOpen={mobileMenuOpen} onClick={toggleMobileMenu} />
        
        <MobileMenu isOpen={mobileMenuOpen}>
          <CloseButton onClick={toggleMobileMenu}>
            <FaTimes />
          </CloseButton>
          
          <Logo to="/" onClick={() => setMobileMenuOpen(false)}>
            Learn<span>Hub</span>
          </Logo>
          
          <MobileNavLinks>
            <MobileNavLink to="/courses" onClick={() => setMobileMenuOpen(false)}>
              Courses
            </MobileNavLink>
            <MobileNavLink to="/instructors" onClick={() => setMobileMenuOpen(false)}>
              Instructors
            </MobileNavLink>
            <MobileNavLink to="/about" onClick={() => setMobileMenuOpen(false)}>
              About
            </MobileNavLink>
            <MobileNavLink to="/admin" onClick={() => setMobileMenuOpen(false)}>
              Admin Panel
            </MobileNavLink>
            
            {isAuthenticated ? (
              <>
                <MobileNavLink to="/dashboard" onClick={() => setMobileMenuOpen(false)}>
                  Dashboard
                </MobileNavLink>
                <MobileNavLink to="/profile" onClick={() => setMobileMenuOpen(false)}>
                  Profile
                </MobileNavLink>
                <Button 
                  variant="primary"
                  onClick={() => {
                    handleLogout();
                    setMobileMenuOpen(false);
                  }}
                >
                  Log Out
                </Button>
              </>
            ) : (
              <>
                <MobileNavLink to="/signin" onClick={() => setMobileMenuOpen(false)}>
                  Sign In
                </MobileNavLink>
                <Button 
                  as={Link} 
                  to="/signup" 
                  variant="primary"
                  onClick={() => setMobileMenuOpen(false)}
                >
                  Sign Up
                </Button>
              </>
            )}
          </MobileNavLinks>
        </MobileMenu>
      </HeaderContainer>
    </HeaderWrapper>
  );
};

export default Header;
