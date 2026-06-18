import React, { useState } from 'react';
import styled from 'styled-components';
import { useNavigate, Link } from 'react-router-dom';
import { FcGoogle } from 'react-icons/fc';
import { AiOutlineEye, AiOutlineEyeInvisible } from 'react-icons/ai';
import { loginUser, signInWithGoogle, resetPassword } from '../firebase/authService';
import { 
  Container,
  Card,
  Title,
  Text,
  Button,
  Input,
  FormGroup,
  Label,
  ErrorMessage,
  Flex,
  Divider,
  SocialButton
} from '../components/ui';

const SignInContainer = styled(Container)`
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem 1rem;
`;

const SignInCard = styled(Card)`
  max-width: 500px;
  width: 100%;
`;

const PasswordWrapper = styled.div`
  position: relative;
  width: 100%;
`;

const PasswordToggle = styled.button`
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  background: none;
  border: none;
  cursor: pointer;
  color: ${({ theme }) => theme.colors.gray};
  font-size: 1.2rem;
  display: flex;
  align-items: center;
  justify-content: center;
`;

const ForgotPassword = styled.button`
  background: none;
  border: none;
  cursor: pointer;
  color: ${({ theme }) => theme.colors.primary};
  font-size: ${({ theme }) => theme.fontSizes.sm};
  text-align: right;
  display: block;
  margin-left: auto;
  margin-top: -0.5rem;
  margin-bottom: 1rem;
  
  &:hover {
    text-decoration: underline;
  }
`;

const LoginIllustration = styled.div`
  display: none;

  @media (min-width: ${({ theme }) => theme.breakpoints.md}) {
    display: block;
    max-width: 400px;
    width: 100%;
    
    img {
      width: 100%;
      height: auto;
    }
  }
`;

const LoginLayout = styled(Flex)`
  gap: 2rem;
  align-items: center;
  justify-content: center;
`;

const SuccessMessage = styled.div`
  background-color: ${({ theme }) => theme.colors.success}20;
  color: ${({ theme }) => theme.colors.success};
  padding: 0.75rem;
  border-radius: ${({ theme }) => theme.borderRadius.md};
  margin-bottom: 1rem;
  text-align: center;
`;

const SignIn = () => {
  const [formData, setFormData] = useState({
    email: '',
    password: ''
  });
  const [showPassword, setShowPassword] = useState(false);
  const [errors, setErrors] = useState({});
  const [loading, setLoading] = useState(false);
  const [resetSent, setResetSent] = useState(false);
  const navigate = useNavigate();

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData({ ...formData, [name]: value });
    // Clear error when user types
    if (errors[name]) {
      setErrors({ ...errors, [name]: '' });
    }
  };

  const validateForm = () => {
    const newErrors = {};
    if (!formData.email.trim()) newErrors.email = 'Email is required';
    else if (!/\S+@\S+\.\S+/.test(formData.email)) newErrors.email = 'Email is invalid';
    
    if (!formData.password) newErrors.password = 'Password is required';
    
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    if (!validateForm()) return;
    
    setLoading(true);
    try {
      const { success, error } = await loginUser(
        formData.email, 
        formData.password
      );
      
      if (success) {
        navigate('/dashboard');
      } else {
        setErrors({ submit: error });
      }
    } catch (error) {
      setErrors({ submit: error.message });
    } finally {
      setLoading(false);
    }
  };

  const handleGoogleSignIn = async () => {
    setLoading(true);
    try {
      const { success, error } = await signInWithGoogle();
      
      if (success) {
        navigate('/dashboard');
      } else {
        setErrors({ submit: error });
      }
    } catch (error) {
      setErrors({ submit: error.message });
    } finally {
      setLoading(false);
    }
  };

  const handlePasswordReset = async () => {
    if (!formData.email.trim()) {
      setErrors({ email: 'Please enter your email to reset password' });
      return;
    }
    
    if (!/\S+@\S+\.\S+/.test(formData.email)) {
      setErrors({ email: 'Please enter a valid email address' });
      return;
    }
    
    setLoading(true);
    try {
      const { success, error } = await resetPassword(formData.email);
      
      if (success) {
        setResetSent(true);
        setErrors({});
      } else {
        setErrors({ submit: error });
      }
    } catch (error) {
      setErrors({ submit: error.message });
    } finally {
      setLoading(false);
    }
  };

  return (
    <SignInContainer>
      <LoginLayout direction="row" wrap="wrap">
        <LoginIllustration>
          <img src="/images/login-illustration.svg" alt="Sign In" />
        </LoginIllustration>
        
        <SignInCard>
          <Title center>Welcome Back</Title>
          <Text center color="gray">Sign in to continue to your account</Text>
          
          {errors.submit && (
            <ErrorMessage>{errors.submit}</ErrorMessage>
          )}
          
          {resetSent && (
            <SuccessMessage>
              Password reset link has been sent to your email address.
            </SuccessMessage>
          )}
          
          <form onSubmit={handleSubmit}>
            <FormGroup>
              <Label htmlFor="email">Email Address</Label>
              <Input
                type="email"
                id="email"
                name="email"
                placeholder="Enter your email"
                value={formData.email}
                onChange={handleChange}
                error={errors.email}
              />
              {errors.email && <ErrorMessage>{errors.email}</ErrorMessage>}
            </FormGroup>
            
            <FormGroup>
              <Label htmlFor="password">Password</Label>
              <PasswordWrapper>
                <Input
                  type={showPassword ? "text" : "password"}
                  id="password"
                  name="password"
                  placeholder="Enter your password"
                  value={formData.password}
                  onChange={handleChange}
                  error={errors.password}
                />
                <PasswordToggle 
                  type="button" 
                  onClick={() => setShowPassword(!showPassword)}
                  aria-label={showPassword ? "Hide password" : "Show password"}
                >
                  {showPassword ? <AiOutlineEyeInvisible /> : <AiOutlineEye />}
                </PasswordToggle>
              </PasswordWrapper>
              {errors.password && <ErrorMessage>{errors.password}</ErrorMessage>}
            </FormGroup>
            
            <ForgotPassword type="button" onClick={handlePasswordReset}>
              Forgot Password?
            </ForgotPassword>
            
            <Button 
              type="submit" 
              variant="primary" 
              fullWidth 
              disabled={loading}
            >
              {loading ? 'Signing In...' : 'Sign In'}
            </Button>
          </form>
          
          <Divider />
          
          <SocialButton onClick={handleGoogleSignIn} disabled={loading}>
            <FcGoogle size={24} /> Continue with Google
          </SocialButton>
          
          <Text center style={{ marginTop: '1.5rem' }}>
            Don't have an account? <Link to="/signup">Sign Up</Link>
          </Text>
        </SignInCard>
      </LoginLayout>
    </SignInContainer>
  );
};

export default SignIn;
