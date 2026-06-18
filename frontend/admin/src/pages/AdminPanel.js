import React, { useEffect, useState } from 'react';
import styled from 'styled-components';
import { db } from '../firebase/firebaseConfig';
import { collection, getDocs, addDoc, Timestamp, query } from 'firebase/firestore';
import { Container, Title, Text, Button, Card, Flex } from '../components/ui';
import { FaUsers, FaBook, FaPlusCircle } from 'react-icons/fa';

const AdminPanelContainer = styled(Container)`
  padding-top: 2rem;
  padding-bottom: 4rem;
`;

const TabsContainer = styled.div`
  display: flex;
  margin-bottom: 2rem;
  border-bottom: 1px solid ${({ theme }) => theme.colors.gray};
`;

const Tab = styled.button`
  padding: 1rem 1.5rem;
  background: transparent;
  border: none;
  font-size: ${({ theme }) => theme.fontSizes.md};
  font-weight: ${({ active }) => (active ? 'bold' : 'normal')};
  color: ${({ active, theme }) => (active ? theme.colors.primary : theme.colors.text)};
  cursor: pointer;
  position: relative;

  &::after {
    content: '';
    position: absolute;
    bottom: -1px;
    left: 0;
    width: 100%;
    height: 3px;
    background: ${({ active, theme }) => (active ? theme.colors.primary : 'transparent')};
  }
`;

const AddCourseForm = styled.form`
  margin-top: 2rem;
  padding: 2rem;
  background: ${({ theme }) => theme.colors.white};
  border-radius: ${({ theme }) => theme.borderRadius.md};
  box-shadow: ${({ theme }) => theme.shadows.md};

  input, textarea {
    width: 100%;
    margin-bottom: 1rem;
    padding: 0.75rem;
    border: 1px solid ${({ theme }) => theme.colors.gray};
    border-radius: ${({ theme }) => theme.borderRadius.sm};
  }

  button {
    margin-top: 1rem;
  }
`;

const CourseCard = styled(Card)`
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

const UserCard = styled(Card)`
  padding: 1.5rem;
  margin-bottom: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  transition: transform 0.2s;
  
  &:hover {
    transform: translateY(-5px);
    box-shadow: ${({ theme }) => theme.shadows.lg};
  }
`;

const CourseGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(1, 1fr);
  gap: 2rem;
  
  @media (min-width: ${({ theme }) => theme.breakpoints.md}) {
    grid-template-columns: repeat(2, 1fr);
  }
  
  @media (min-width: ${({ theme }) => theme.breakpoints.lg}) {
    grid-template-columns: repeat(3, 1fr);
  }
`;

const PinContainer = styled.div`
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100vh;
  background: ${({ theme }) => theme.colors.lightGray}50;
`;

const PinInput = styled.input`
  width: 200px;
  padding: 0.75rem;
  margin-bottom: 1rem;
  border: 1px solid ${({ theme }) => theme.colors.gray};
  border-radius: ${({ theme }) => theme.borderRadius.sm};
  font-size: ${({ theme }) => theme.fontSizes.md};
  text-align: center;
`;

const AdminPanel = () => {
  const [users, setUsers] = useState([]);
  const [courses, setCourses] = useState([]);
  const [newCourse, setNewCourse] = useState({ title: '', description: '', instructor: '', duration: '', cover: '' });
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [pin, setPin] = useState('');
  const [activeTab, setActiveTab] = useState('users');

  useEffect(() => {
    const fetchUsersWithEnrollments = async () => {
      const usersSnapshot = await getDocs(collection(db, 'users'));
      const usersData = await Promise.all(
        usersSnapshot.docs.map(async (doc) => {
          const userData = { id: doc.id, ...doc.data() };
          const enrollmentQuery = query(collection(db, `users/${doc.id}/enrollment`));
          const enrollmentSnapshot = await getDocs(enrollmentQuery);
          userData.enrolledCourses = enrollmentSnapshot.docs.map((enrollmentDoc) => ({
            id: enrollmentDoc.id,
            ...enrollmentDoc.data(),
          }));
          return userData;
        })
      );
      setUsers(usersData);
    };

    const fetchCourses = async () => {
      const coursesSnapshot = await getDocs(collection(db, 'courses'));
      const coursesData = coursesSnapshot.docs.map(doc => ({ id: doc.id, ...doc.data() }));
      setCourses(coursesData);
    };

    const fetchData = async () => {
      await fetchUsersWithEnrollments();
      fetchCourses();
    };

    fetchData();
  }, []);

  const handleAddCourse = async (e) => {
    e.preventDefault();
    try {
      const formattedCourse = {
        ...newCourse,
        instructor: newCourse.instructor.split(',').map((name) => name.trim()), // Convert instructor string to an array
      };
      await addDoc(collection(db, 'courses'), formattedCourse);
      setCourses([...courses, formattedCourse]);
      setNewCourse({ title: '', description: '', instructor: '', duration: '', cover: '' });
    } catch (error) {
      console.error('Error adding course:', error);
    }
  };

  const handlePinSubmit = (e) => {
    e.preventDefault();
    if (pin === '1234') {
      setIsAuthenticated(true);
    } else {
      alert('Invalid PIN');
    }
  };

  const formatTimestamp = (timestamp) => {
    if (timestamp instanceof Timestamp) {
      return timestamp.toDate().toLocaleString();
    }
    return timestamp;
  };

  const renderTabContent = () => {
    switch (activeTab) {
      case 'users':
        return (
          <>
            <Flex justify="space-between" align="center" style={{ marginBottom: '1rem' }}>
              <Title size="md"><FaUsers /> Users ({users.length})</Title>
            </Flex>
            {users.map(user => (
              <UserCard key={user.id}>
                <Title size="sm">{user.name}</Title>
                <Text><strong>Email:</strong> {user.email}</Text>
                <Text><strong>Last Login:</strong> {formatTimestamp(user.lastLogin)}</Text>
                <Text><strong>Enrolled Courses:</strong> {user.enrolledCourses ? user.enrolledCourses.length : 0}</Text>
              </UserCard>
            ))}
          </>
        );
      case 'courses':
        return (
          <>
            <Flex justify="space-between" align="center" style={{ marginBottom: '1rem' }}>
              <Title size="md"><FaBook /> Courses ({courses.length})</Title>
            </Flex>
            <CourseGrid>
              {courses.map(course => (
                <CourseCard key={course.id}>
                  <img src={course.cover || '/default-course-cover.jpg'} alt={course.title} />
                  <div className="content">
                    <Title size="sm">{course.title}</Title>
                    <Text><strong>Instructor:</strong> {Array.isArray(course.instructor) 
                      ? course.instructor.join(', ') 
                      : course.instructor}</Text>
                    <Text><strong>Duration:</strong> {course.duration}</Text>
                    <Text>{course.description}</Text>
                  </div>
                </CourseCard>
              ))}
            </CourseGrid>
          </>
        );
      case 'addCourse':
        return (
          <>
            <Flex justify="space-between" align="center" style={{ marginBottom: '1rem' }}>
              <Title size="md"><FaPlusCircle /> Add New Course</Title>
            </Flex>
            <Card style={{ padding: '2rem' }}>
              <AddCourseForm onSubmit={handleAddCourse}>
                <input
                  type="text"
                  placeholder="Course Title"
                  value={newCourse.title}
                  onChange={(e) => setNewCourse({ ...newCourse, title: e.target.value })}
                  required
                />
                <textarea
                  placeholder="Course Description"
                  value={newCourse.description}
                  onChange={(e) => setNewCourse({ ...newCourse, description: e.target.value })}
                  required
                ></textarea>
                <input
                  type="text"
                  placeholder="Instructor (comma separated for multiple)"
                  value={newCourse.instructor}
                  onChange={(e) => setNewCourse({ ...newCourse, instructor: e.target.value })}
                  required
                />
                <input
                  type="text"
                  placeholder="Duration (e.g., 8 weeks)"
                  value={newCourse.duration}
                  onChange={(e) => setNewCourse({ ...newCourse, duration: e.target.value })}
                  required
                />
                <input
                  type="text"
                  placeholder="Cover Image URL"
                  value={newCourse.cover}
                  onChange={(e) => setNewCourse({ ...newCourse, cover: e.target.value })}
                />
                <Button type="submit" variant="primary">Add Course</Button>
              </AddCourseForm>
            </Card>
          </>
        );
      default:
        return null;
    }
  };

  if (!isAuthenticated) {
    return (
      <PinContainer>
        <Title size="lg">Enter Admin PIN</Title>
        <form onSubmit={handlePinSubmit}>
          <PinInput
            type="password"
            placeholder="Enter PIN"
            value={pin}
            onChange={(e) => setPin(e.target.value)}
          />
          <Button type="submit" variant="primary">Submit</Button>
        </form>
      </PinContainer>
    );
  }

  return (
    <AdminPanelContainer>
      <Title size="lg">Admin Panel</Title>
      
      <TabsContainer>
        <Tab 
          active={activeTab === 'users'} 
          onClick={() => setActiveTab('users')}
        >
          <FaUsers /> Users
        </Tab>
        <Tab 
          active={activeTab === 'courses'} 
          onClick={() => setActiveTab('courses')}
        >
          <FaBook /> Courses
        </Tab>
        <Tab 
          active={activeTab === 'addCourse'} 
          onClick={() => setActiveTab('addCourse')}
        >
          <FaPlusCircle /> Add Course
        </Tab>
      </TabsContainer>
      
      {renderTabContent()}
    </AdminPanelContainer>
  );
};

export default AdminPanel;
