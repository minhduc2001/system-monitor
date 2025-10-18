import { BrowserRouter } from "react-router-dom";
import ErrorBoundary from "./components/ErrorBoundary";
import { ReactQueryProvider } from "./providers/ReactQueryProvider";
import { AppRoutes } from "@/router/routes";

function App() {
  return (
    <ErrorBoundary>
      <BrowserRouter>
        <ReactQueryProvider>
          <AppRoutes />
        </ReactQueryProvider>
      </BrowserRouter>
    </ErrorBoundary>
  );
}

export default App;
