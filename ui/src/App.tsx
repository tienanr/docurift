import React, { useEffect, useState } from 'react';
import { Container, Grid, CssBaseline, ThemeProvider, createTheme } from '@mui/material';
import EndpointList from './components/EndpointList';
import EndpointDetail from './components/EndpointDetail';
import { AnalyzerData } from './types/analyzer';

const theme = createTheme({
  palette: {
    mode: 'light',
    primary: {
      main: '#1976d2',
    },
  },
});

function App() {
  const [data, setData] = useState<AnalyzerData>({});
  const [selectedEndpoint, setSelectedEndpoint] = useState<string | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const response = await fetch('http://localhost:9877/api/analyzer');
        const jsonData = await response.json();
        console.log('Received data from analyzer:', jsonData);
        setData(jsonData);
      } catch (error) {
        console.error('Error fetching analyzer data:', error);
      }
    };

    fetchData();
    const interval = setInterval(fetchData, 5000); // Refresh every 5 seconds

    return () => clearInterval(interval);
  }, []);

  // Add debug logging for data state
  useEffect(() => {
    console.log('Current data state:', data);
  }, [data]);

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Container maxWidth="xl" sx={{ py: 4 }}>
        <Grid container spacing={3} sx={{ height: 'calc(100vh - 64px)' }}>
          <Grid item xs={12} md={4}>
            <EndpointList
              data={data}
              onSelectEndpoint={setSelectedEndpoint}
            />
          </Grid>
          <Grid item xs={12} md={8}>
            {selectedEndpoint && data[selectedEndpoint] ? (
              <EndpointDetail endpoint={data[selectedEndpoint]} />
            ) : (
              <div>Select an endpoint to view details</div>
            )}
          </Grid>
        </Grid>
      </Container>
    </ThemeProvider>
  );
}

export default App; 