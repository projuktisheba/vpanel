import {
  BrowserRouter as Router,
  Routes,
  Route,
  Navigate,
} from "react-router-dom";
import SignIn from "./pages/AuthPages/SignIn";
import SignUp from "./pages/AuthPages/SignUp";
import NotFound from "./pages/OtherPage/NotFound";
import UserProfiles from "./pages/UserProfiles";
import AppLayout from "./layout/AppLayout";
import { ScrollToTop } from "./components/common/ScrollToTop";
import { ProtectedRoute } from "./components/protectedRoute";
import { AuthProvider } from "./context/AuthContext";
import Home from "./pages/Dashboard/Home";
import MySQL from "./pages/Database/MySQL";
import PostgreSQL from "./pages/Database/PostgreSQL";
import PHP from "./pages/AppManager/PHP";
import PreBuildSite from "./pages/AppManager/PreBuildSite";
import Domain from "./pages/Domain/DomainManager";
import WordpressSiteBuilder from "./pages/AppManager/Wordpress";

export default function App() {
  return (
    <Router>
      <AuthProvider>
        <ScrollToTop />
        <Routes>
          {/* ==================== Protected Dashboard Layout ==================== */}
          <Route
            element={
              <ProtectedRoute>
                <AppLayout />
              </ProtectedRoute>
            }
          >
            <Route index path="/" element={<Home />} />
            <Route path="/profile" element={<UserProfiles />} />

            {/* APP Manager */}
            <Route path="/wordpress" element={<WordpressSiteBuilder />} />
            <Route path="/php" element={<PHP />} />
            <Route path="/pre-build" element={<PreBuildSite />} />

            {/* Database Manager */}
            <Route path="/mysql" element={<MySQL />} />
            <Route path="/postgresql" element={<PostgreSQL />} />
            {/* Domain Manager */}
            <Route path="/domain" element={<Domain />} />
          </Route>

          {/* ==================== Auth Pages ==================== */}
          <Route path="/signin" element={<SignIn />} />
          <Route path="/signup" element={<SignUp />} />

          {/* Redirect root → dashboard */}
          <Route path="/" element={<Navigate to="/" replace />} />

          {/* ==================== 404 Fallback ==================== */}
          <Route path="*" element={<NotFound />} />
        </Routes>
      </AuthProvider>
    </Router>
  );
}
