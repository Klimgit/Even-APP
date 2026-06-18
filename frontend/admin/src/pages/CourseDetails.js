import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import styled from 'styled-components';
import { db } from '../firebase/firebaseConfig';
import { doc, getDoc } from 'firebase/firestore';
import { Container, Title, Text, Button, Card, Flex } from '../components/ui';
import { FaArrowLeft } from 'react-icons/fa';

const CourseDetailsContainer = styled(Container)`
  padding-top: 2rem;
  padding-bottom: 4rem;
`;

const CoverImage = styled.img`
  width: 100%;
  height: 300px;
  object-fit: cover;
  border-radius: ${({ theme }) => theme.borderRadius.md};
  margin-bottom: 2rem;
`;

const Timeline = styled.div`
  margin-top: 2rem;
  border-left: 4px solid ${({ theme }) => theme.colors.primary};
  padding-left: 1rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
`;

const TimelineItem = styled.div`
  position: relative;
  padding-left: 1.5rem;
  &::before {
    content: '';
    position: absolute;
    left: -1.5rem;
    top: 0.5rem;
    width: 1rem;
    height: 1rem;
    background-color: ${({ theme }) => theme.colors.primary};
    border-radius: 50%;
  }
`;

const DetailsCard = styled(Card)`
  padding: 2rem;
  margin-top: 2rem;
  background: ${({ theme }) => theme.colors.white};
  color: ${({ theme }) => theme.colors.text};
  box-shadow: ${({ theme }) => theme.shadows.md};
  border-radius: ${({ theme }) => theme.borderRadius.md};
`;

const RelatedCoursesSection = styled.div`
  margin-top: 3rem;
  padding: 2rem;
  background: ${({ theme }) => theme.colors.lightGray}50;
  border-radius: ${({ theme }) => theme.borderRadius.md};
  box-shadow: ${({ theme }) => theme.shadows.sm};
`;

const CourseDetails = () => {
  const { courseId } = useParams();
  const navigate = useNavigate();
  const [course, setCourse] = useState(null);
  const [relatedCourses, setRelatedCourses] = useState([]);

  useEffect(() => {
    const fetchCourse = async () => {
      try {
        const courseRef = doc(db, 'courses', courseId);
        const courseSnap = await getDoc(courseRef);
        if (courseSnap.exists()) {
          setCourse({ id: courseSnap.id, ...courseSnap.data() });
        } else {
          console.error('Course not found');
        }
      } catch (error) {
        console.error('Error fetching course details:', error);
      }
    };

    const fetchRelatedCourses = async () => {
      // Fetch related courses logic here
      // For now, we will simulate with static data
      const related = [
        { id: '1', title: 'Related Course 1', description: 'Description for related course 1' },
        { id: '2', title: 'Related Course 2', description: 'Description for related course 2' },
        { id: '3', title: 'Related Course 3', description: 'Description for related course 3' },
      ];
      setRelatedCourses(related);
    };

    fetchCourse();
    fetchRelatedCourses();
  }, [courseId]);

  if (!course) {
    return <Text>Loading...</Text>;
  }

  return (
    <CourseDetailsContainer>
      <Button variant="outline" onClick={() => navigate(-1)}>
        <FaArrowLeft /> Back
      </Button>

      <CoverImage src={course.cover || '/default-course-cover.jpg'} alt={course.title} />

      <DetailsCard>
        <Title>{course.title}</Title>
        <Text>Instructor: {course.instructor || 'Unknown'}</Text>
        <Text>Duration: {course.duration || 'N/A'}</Text>
        <Text style={{ marginTop: '1rem' }}>
          This course is designed to provide you with the best learning experience. You will gain
          valuable knowledge and skills to excel in your field. Enroll now and start your journey
          towards success!
        </Text>
      </DetailsCard>

      <Timeline>
        <Title size="md">Course Timeline</Title>
        <TimelineItem>
          <Text>Introduction to the course</Text>
        </TimelineItem>
        <TimelineItem>
          <Text>Core concepts and fundamentals</Text>
        </TimelineItem>
        <TimelineItem>
          <Text>Hands-on projects and assignments</Text>
        </TimelineItem>
        <TimelineItem>
          <Text>Final assessment and certification</Text>
        </TimelineItem>
      </Timeline>

      <RelatedCoursesSection>
        <Title size="md">Related Courses</Title>
        <div>
          {relatedCourses.map(course => (
            <Card key={course.id} style={{ marginBottom: '1rem', padding: '1rem' }}>
              <Title size="sm">{course.title}</Title>
              <Text>{course.description}</Text>
              <Button
                variant="primary"
                style={{ marginTop: '1rem' }}
                onClick={() => navigate(`/courses/${course.id}`)}
              >
                View Details
              </Button>
            </Card>
          ))}
        </div>
      </RelatedCoursesSection>
    </CourseDetailsContainer>
  );
};

export default CourseDetails;
