import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import styled from 'styled-components';
import { Container, Title, Text, Button } from '../components/ui';
import { FaArrowLeft } from 'react-icons/fa';
import { doc, getDoc } from 'firebase/firestore';
import { db } from '../firebase/firebaseConfig'; // Corrected the import path

const EnrolledCourseContainer = styled(Container)`
  padding-top: 2rem;
  padding-bottom: 4rem;
`;

const VideoEmbed = styled.div`
  position: relative;
  padding-bottom: 56.25%;
  height: 0;
  overflow: hidden;
  iframe {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
  }
`;

const NotesSection = styled.div`
  margin-top: 2rem;
  padding: 1.5rem;
  background: ${({ theme }) => theme.colors.white};
  border-radius: ${({ theme }) => theme.borderRadius.md};
  box-shadow: ${({ theme }) => theme.shadows.sm};
`;

const NotesTextarea = styled.textarea`
  width: 100%;
  height: 100px;
  padding: 0.75rem;
  border: 1px solid ${({ theme }) => theme.colors.gray};
  border-radius: ${({ theme }) => theme.borderRadius.sm};
  font-size: ${({ theme }) => theme.fontSizes.md};
  box-shadow: ${({ theme }) => theme.shadows.sm};
`;

const EnrolledCourseDetails = () => {
  const { courseId } = useParams();
  const navigate = useNavigate();
  const [course, setCourse] = useState(null);

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

    fetchCourse();
  }, [courseId]);

  if (!course) {
    return <Text>Loading...</Text>;
  }

  return (
    <EnrolledCourseContainer>
      <Button variant="outline" onClick={() => navigate(-1)}>
        <FaArrowLeft /> Back
      </Button>

      <Title>{course.title}</Title>
      <Text>Instructor: {course.instructor || 'Unknown'}</Text>
      <Text>Duration: {course.duration || 'N/A'}</Text>

      <VideoEmbed>
        <iframe
          src="https://www.youtube.com/embed/dQw4w9WgXcQ"
          title="Course Video"
          frameBorder="0"
          allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
          allowFullScreen
        ></iframe>
      </VideoEmbed>

      <NotesSection>
        <Title size="md">Notes</Title>
        <NotesTextarea placeholder="Write your notes here..." />
      </NotesSection>
    </EnrolledCourseContainer>
  );
};

export default EnrolledCourseDetails;
