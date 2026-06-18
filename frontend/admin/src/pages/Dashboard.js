import React, { useEffect, useState } from 'react';
import styled from 'styled-components';
import { useAuth } from '../context/AuthContext';
import { logoutUser } from '../firebase/authService';
import { useNavigate } from 'react-router-dom';
import { Container, Title, Button, Card, Flex, Text } from '../components/ui';
import { FaBook, FaChalkboardTeacher, FaCertificate, FaUser, FaSearch } from 'react-icons/fa';
import { db } from '../firebase/firebaseConfig';
import { collection, getDocs, query } from 'firebase/firestore';

const DashboardContainer = styled(Container)`
  padding-top: 2rem;
  padding-bottom: 4rem;
  background: ${({ theme }) => theme.colors.lightGray}20;
  border-radius: ${({ theme }) => theme.borderRadius.md};
  box-shadow: ${({ theme }) => theme.shadows.sm};
`;

const WelcomeCard = styled(Card)`
  background: linear-gradient(135deg, ${({ theme }) => theme.colors.gradientStart}, ${({ theme }) => theme.colors.gradientEnd});
  color: white;
  margin-bottom: 2rem;
  padding: 2rem;
`;

const StatsGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(1, 1fr);
  gap: 1.5rem;
  margin-bottom: 2rem;
  
  @media (min-width: ${({ theme }) => theme.breakpoints.sm}) {
    grid-template-columns: repeat(2, 1fr);
  }
  
  @media (min-width: ${({ theme }) => theme.breakpoints.lg}) {
    grid-template-columns: repeat(4, 1fr);
  }
`;

const StatCard = styled(Card)`
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  transition: ${({ theme }) => theme.transitions.medium};
  border-top: 4px solid ${({ color, theme }) => theme.colors[color] || theme.colors.primary};
  background: ${({ theme }) => theme.colors.white};
  color: ${({ theme }) => theme.colors.text};
  box-shadow: ${({ theme }) => theme.shadows.sm};

  &:hover {
    transform: translateY(-5px);
  }

  svg {
    color: ${({ color, theme }) => theme.colors[color] || theme.colors.primary};
    font-size: 2.5rem;
    margin-bottom: 1rem;
  }

  h3 {
    margin-bottom: 0.25rem;
    font-size: ${({ theme }) => theme.fontSizes.xxl};
  }
`;

const RecentCoursesSection = styled.div`
  margin-top: 2rem;
`;

const EmptyState = styled.div`
  text-align: center;
  padding: 3rem;
  background-color: ${({ theme }) => theme.colors.lightGray}50;
  border-radius: ${({ theme }) => theme.borderRadius.lg};
  
  svg {
    font-size: 3rem;
    color: ${({ theme }) => theme.colors.gray};
    margin-bottom: 1rem;
  }
`;

const AvailableCoursesGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(1, 1fr);
  gap: 1.5rem;
  margin-top: 2rem;

  @media (min-width: ${({ theme }) => theme.breakpoints.sm}) {
    grid-template-columns: repeat(2, 1fr);
  }

  @media (min-width: ${({ theme }) => theme.breakpoints.lg}) {
    grid-template-columns: repeat(3, 1fr);
  }
`;

const AvailableCourseCard = styled(Card)`
  padding: 0;
  overflow: hidden;
  
  img {
    width: 100%;
    height: 150px;
    object-fit: cover;
  }
  
  .content {
    padding: 1rem;
  }
`;

const EnrolledCoursesSection = styled.div`
  margin-bottom: 3rem;
  padding: 2rem;
  background: ${({ theme }) => theme.colors.white};
  border-radius: ${({ theme }) => theme.borderRadius.md};
  box-shadow: ${({ theme }) => theme.shadows.md};
`;

const EnrolledCourseCard = styled(Card)`
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  padding: 1rem;
  box-shadow: ${({ theme }) => theme.shadows.md};

  img {
    width: 120px;
    height: 80px;
    object-fit: cover;
    border-radius: ${({ theme }) => theme.borderRadius.sm};
  }

  .content {
    flex: 1;
  }
`;

const SearchBarContainer = styled.div`
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 1.5rem;
  padding: 0.5rem;
  background: ${({ theme }) => theme.colors.white};
  border: 1px solid ${({ theme }) => theme.colors.gray};
  border-radius: ${({ theme }) => theme.borderRadius.sm};
  box-shadow: ${({ theme }) => theme.shadows.sm};
`;

const SearchIcon = styled(FaSearch)`
  color: ${({ theme }) => theme.colors.gray};
  font-size: 1.25rem;
`;

const StyledSearchInput = styled.input`
  flex: 1;
  border: none;
  outline: none;
  font-size: ${({ theme }) => theme.fontSizes.md};
  color: ${({ theme }) => theme.colors.text};
  &::placeholder {
    color: ${({ theme }) => theme.colors.gray};
  }
`;

const ProgressBar = styled.div`
  width: 100%;
  height: 8px;
  background: ${({ theme }) => theme.colors.lightGray};
  border-radius: ${({ theme }) => theme.borderRadius.sm};
  overflow: hidden;
  margin-top: 0.5rem;
  div {
    height: 100%;
    width: ${({ progress }) => progress || 0}%;
    background: ${({ theme }) => theme.colors.primary};
  }
`;

const Dashboard = () => {
  const { currentUser, userData } = useAuth();
  const navigate = useNavigate();
  const [enrolledCourses, setEnrolledCourses] = useState([]);
  const [availableCourses, setAvailableCourses] = useState([]);

  useEffect(() => {
    const fetchCourses = async () => {
      try {
        // Fetch all available courses
        const coursesQuery = query(collection(db, 'courses'));
        const coursesSnapshot = await getDocs(coursesQuery);
        const courses = coursesSnapshot.docs.map(doc => ({ id: doc.id, ...doc.data() }));
        setAvailableCourses(courses);

        // Fetch enrolled courses for the current user
        if (currentUser) {
          const enrollmentQuery = query(
            collection(db, `users/${currentUser.uid}/enrollment`)
          );
          const enrollmentSnapshot = await getDocs(enrollmentQuery);
          const enrollments = enrollmentSnapshot.docs.map(doc => ({ id: doc.id, ...doc.data() }));
          setEnrolledCourses(enrollments);
        }
      } catch (error) {
        console.error('Error fetching courses:', error);
      }
    };

    fetchCourses();
  }, [currentUser]);

  const handleLogout = async () => {
    try {
      await logoutUser();
      navigate('/signin');
    } catch (error) {
      console.error('Error logging out:', error);
    }
  };

  return (
    <DashboardContainer>
      <WelcomeCard>
        <Flex justify="space-between" align="center" wrap="wrap">
          <div>
            <Title color="white">Welcome, {userData?.fullName || currentUser?.displayName || 'Student'}</Title>
            <Text color="white" style={{ marginBottom: 0 }}>
              {userData?.role === 'admin' ? 'Administrator' : 
                userData?.role === 'instructor' ? 'Instructor' : 'Student'}
            </Text>
          </div>
          <Button variant="outline" onClick={handleLogout}>
            Log Out
          </Button>
        </Flex>
      </WelcomeCard>

      <StatsGrid>
        <StatCard color="primary">
          <FaBook />
          <h3>{enrolledCourses.length}</h3>
          <Text>Enrolled Courses</Text>
        </StatCard>

        <StatCard color="secondary">
          <FaChalkboardTeacher />
          <h3>0</h3>
          <Text>Completed Lessons</Text>
        </StatCard>

        <StatCard color="accent">
          <FaCertificate />
          <h3>0</h3>
          <Text>Certificates</Text>
        </StatCard>

        <StatCard color="success">
          <FaUser />
          <h3>1</h3>
          <Text>Days Active</Text>
        </StatCard>
      </StatsGrid>

      <EnrolledCoursesSection>
        <Title size="lg">Enrolled Courses</Title>
        <SearchBarContainer>
          <SearchIcon />
          <StyledSearchInput placeholder="Search enrolled courses..." />
        </SearchBarContainer>
        {enrolledCourses.length > 0 ? (
          <div>
            {enrolledCourses.map(course => (
              <EnrolledCourseCard key={course.id}>
                <img src={course.cover || '/default-course-cover.jpg'} alt={course.title} />
                <div className="content">
                  <Title size="sm">{course.title}</Title>
                  <Text>{course.description}</Text>
                  <ProgressBar progress={course.progress || 0}>
                    <div></div>
                  </ProgressBar>
                  <Button variant="primary" style={{ marginTop: '1rem' }} onClick={() => navigate(`/enrolled-courses/${course.id}`)}>
                    View Course
                  </Button>
                </div>
              </EnrolledCourseCard>
            ))}
          </div>
        ) : (
          <EmptyState>
            <FaBook size={48} />
            <Title size="md">No enrolled courses</Title>
            <Text>You haven't enrolled in any courses yet.</Text>
            <Button variant="primary" style={{ marginTop: '1rem' }} onClick={() => navigate('/courses')}>
              Browse Courses
            </Button>
          </EmptyState>
        )}
      </EnrolledCoursesSection>

      <RecentCoursesSection>
        <Title size="lg">Available Courses</Title>
        {availableCourses.length > 0 ? (
          <AvailableCoursesGrid>
            {availableCourses.map(course => (
              <AvailableCourseCard key={course.id}>
                <img src={course.cover || '/default-course-cover.jpg'} alt={course.title} />
                <div className="content">
                  <Title size="sm">{course.title}</Title>
                  <Text>{course.description}</Text>
                  <Button variant="primary" style={{ marginTop: '1rem' }} onClick={() => navigate(`/courses/${course.id}`)}>
                    View Details
                  </Button>
                </div>
              </AvailableCourseCard>
            ))}
          </AvailableCoursesGrid>
        ) : (
          <EmptyState>
            <FaBook size={48} />
            <Title size="md">No courses available</Title>
            <Text>Check back later for new courses.</Text>
          </EmptyState>
        )}
      </RecentCoursesSection>
    </DashboardContainer>
  );
};

export default Dashboard;
