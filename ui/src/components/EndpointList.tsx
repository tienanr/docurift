import React, { useEffect } from 'react';
import { List, ListItem, ListItemText, Paper, Typography, Box } from '@mui/material';
import { AnalyzerData } from '../types/analyzer';

interface EndpointListProps {
  data: AnalyzerData;
  onSelectEndpoint: (endpoint: string) => void;
}

const EndpointList: React.FC<EndpointListProps> = ({ data, onSelectEndpoint }) => {
  useEffect(() => {
    console.log('EndpointList received data:', data);
    console.log('Number of endpoints:', Object.keys(data).length);
  }, [data]);

  return (
    <Paper elevation={2} sx={{ p: 2, height: '100%', overflow: 'auto' }}>
      <Typography variant="h6" gutterBottom>
        Endpoints ({Object.keys(data).length})
      </Typography>
      <List>
        {Object.entries(data).map(([key, endpoint]) => (
          <ListItem
            key={key}
            button
            onClick={() => onSelectEndpoint(key)}
            sx={{
              borderLeft: '4px solid',
              borderColor: 'primary.main',
              mb: 1,
              '&:hover': {
                backgroundColor: 'action.hover',
              },
            }}
          >
            <ListItemText
              primary={
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                  <Typography
                    component="span"
                    sx={{
                      color: 'primary.main',
                      fontWeight: 'bold',
                      minWidth: '60px',
                    }}
                  >
                    {endpoint.Method}
                  </Typography>
                  <Typography component="span">{endpoint.URL}</Typography>
                </Box>
              }
              secondary={`${Object.keys(endpoint.ResponseStatuses).length} response statuses`}
            />
          </ListItem>
        ))}
      </List>
    </Paper>
  );
};

export default EndpointList; 