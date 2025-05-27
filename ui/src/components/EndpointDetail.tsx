import React from 'react';
import {
  Paper,
  Typography,
  Box,
  Tabs,
  Tab,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
} from '@mui/material';
import { EndpointData } from '../types/analyzer';

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

function TabPanel(props: TabPanelProps) {
  const { children, value, index, ...other } = props;

  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`endpoint-tabpanel-${index}`}
      aria-labelledby={`endpoint-tab-${index}`}
      {...other}
    >
      {value === index && <Box sx={{ p: 3 }}>{children}</Box>}
    </div>
  );
}

interface EndpointDetailProps {
  endpoint: EndpointData;
}

const EndpointDetail: React.FC<EndpointDetailProps> = ({ endpoint }) => {
  const [tabValue, setTabValue] = React.useState(0);

  const handleTabChange = (event: React.SyntheticEvent, newValue: number) => {
    setTabValue(newValue);
  };

  const renderSchemaStore = (store: { Examples: { [key: string]: any[] }; Optional: { [key: string]: boolean } }) => {
    return (
      <TableContainer>
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>Path</TableCell>
              <TableCell>Examples</TableCell>
              <TableCell>Optional</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {Object.entries(store.Examples).map(([path, examples]) => (
              <TableRow key={path}>
                <TableCell>{path}</TableCell>
                <TableCell>
                  {examples.map((example, i) => (
                    <Box key={i} component="pre" sx={{ m: 0, p: 0.5, bgcolor: 'grey.100' }}>
                      {JSON.stringify(example, null, 2)}
                    </Box>
                  ))}
                </TableCell>
                <TableCell>{store.Optional[path] ? 'Yes' : 'No'}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    );
  };

  return (
    <Paper elevation={2} sx={{ height: '100%', overflow: 'auto' }}>
      <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
        <Tabs value={tabValue} onChange={handleTabChange}>
          <Tab label="Request" />
          <Tab label="Responses" />
        </Tabs>
      </Box>

      <TabPanel value={tabValue} index={0}>
        <Typography variant="h6" gutterBottom>
          Request Headers
        </Typography>
        {renderSchemaStore(endpoint.RequestHeaders)}

        <Typography variant="h6" gutterBottom sx={{ mt: 4 }}>
          Request Payload
        </Typography>
        {renderSchemaStore(endpoint.RequestPayload)}

        <Typography variant="h6" gutterBottom sx={{ mt: 4 }}>
          URL Parameters
        </Typography>
        {renderSchemaStore(endpoint.URLParameters)}
      </TabPanel>

      <TabPanel value={tabValue} index={1}>
        {Object.entries(endpoint.ResponseStatuses).map(([status, data]) => (
          <Box key={status} sx={{ mb: 4 }}>
            <Typography variant="h6" gutterBottom>
              Status {status}
            </Typography>

            <Typography variant="subtitle1" gutterBottom>
              Response Headers
            </Typography>
            {renderSchemaStore(data.Headers)}

            <Typography variant="subtitle1" gutterBottom sx={{ mt: 2 }}>
              Response Payload
            </Typography>
            {renderSchemaStore(data.Payload)}
          </Box>
        ))}
      </TabPanel>
    </Paper>
  );
};

export default EndpointDetail; 