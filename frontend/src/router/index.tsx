import type { ReactElement } from 'react';
import { Route, Routes } from 'react-router-dom';
import { RequireAuth } from './guards';
import { Dashboard } from '../pages/Dashboard';
import { Moods } from '../pages/Moods';
import { Assessments } from '../pages/Assessments';
import { Journals } from '../pages/Journals';
import { Profile } from '../pages/Profile';
import { Login } from '../pages/Login';
import { Shell } from '../shell';

export function AppRouter() {
  const secure = (Page: () => ReactElement | null) => (
    <RequireAuth>
      <Shell>
        <Page />
      </Shell>
    </RequireAuth>
  );

  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/dashboard" element={secure(Dashboard)} />
      <Route path="/moods" element={secure(Moods)} />
      <Route path="/assessments" element={secure(Assessments)} />
      <Route path="/journals" element={secure(Journals)} />
      <Route path="/profile" element={secure(Profile)} />
      <Route path="*" element={<RequireAuth><Shell><Dashboard /></Shell></RequireAuth>} />
    </Routes>
  );
}
