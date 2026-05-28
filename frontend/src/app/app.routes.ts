import { Routes } from '@angular/router';
import { DashboardComponent } from './components/dashboard/dashboard';
import { CreatorComponent } from './components/creator/creator';
import { GuestViewComponent } from './components/guest-view/guest-view';

export const routes: Routes = [
  { path: '', component: DashboardComponent },
  { path: 'create', component: CreatorComponent },
  { path: 'edit/:id', component: CreatorComponent },
  { path: 'invitation/:id', component: GuestViewComponent },
  { path: '**', redirectTo: '' }
];
