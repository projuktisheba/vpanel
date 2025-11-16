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
import Videos from "./pages/UiElements/Videos";
import Images from "./pages/UiElements/Images";
import Alerts from "./pages/UiElements/Alerts";
import Badges from "./pages/UiElements/Badges";
import Avatars from "./pages/UiElements/Avatars";
import Buttons from "./pages/UiElements/Buttons";
import LineChart from "./pages/Charts/LineChart";
import BarChart from "./pages/Charts/BarChart";
import Calendar from "./pages/Calendar";
import Blank from "./pages/Blank";
import AppLayout from "./layout/AppLayout";
import { ScrollToTop } from "./components/common/ScrollToTop";
import { ProtectedRoute } from "./components/protectedRoute";
import { AuthProvider } from "./context/AuthContext";
import Home from "./pages/Dashboard/Home";
import MySQL from "./pages/DatabaseMangement/MySQL";
import PostgreSQL from "./pages/DatabaseMangement/PostgreSQL";
import Wordpress from "./pages/AppManager/Wordpress";
import PHP from "./pages/AppManager/PHP";
import PreBuildSite from "./pages/AppManager/PreBuildSite";
import Domain from "./pages/Domain/DomainManager";

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
            <Route path="/calendar" element={<Calendar />} />
            <Route path="/blank" element={<Blank />} />

            {/* APP Manager */}
            <Route path="/wordpress" element={<Wordpress />} />
            <Route path="/php" element={<PHP />} />
            <Route path="/pre-build" element={<PreBuildSite />} />

            {/* Database Manager */}
            <Route path="/mysql" element={<MySQL />} />
            <Route path="/postgresql" element={<PostgreSQL />} />
            {/* Domain Manager */}
            <Route path="/domain" element={<Domain />} />

            {/* UI Elements */}
            <Route path="/alerts" element={<Alerts />} />
            <Route path="/avatars" element={<Avatars />} />
            <Route path="/badge" element={<Badges />} />
            <Route path="/buttons" element={<Buttons />} />
            <Route path="/images" element={<Images />} />
            <Route path="/videos" element={<Videos />} />

            {/* Charts */}
            <Route path="/line-chart" element={<LineChart />} />
            <Route path="/bar-chart" element={<BarChart />} />
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
