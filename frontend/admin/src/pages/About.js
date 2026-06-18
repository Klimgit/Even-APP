import React from 'react';
import styled from 'styled-components';
import { Container, Title, Text } from '../components/ui';

const AboutContainer = styled(Container)`
  padding-top: 2rem;
  padding-bottom: 4rem;
  text-align: center;
`;

const About = () => {
  return (
    <AboutContainer>
      <Title size="lg">About Us</Title>
      <Text>
        Welcome to our e-learning platform! Our mission is to provide high-quality, accessible, and
        engaging educational content to learners around the world. Whether you're looking to develop
        new skills, advance your career, or explore a new hobby, we have courses designed to meet
        your needs.
      </Text>
      <Text>
        Our platform is built with a focus on user experience, offering a seamless and interactive
        learning environment. Join us today and start your journey towards knowledge and growth!
      </Text>
    </AboutContainer>
  );
};

export default About;
