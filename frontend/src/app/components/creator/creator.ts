import { Component, inject, signal, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { Invitation, Item, LayoutConfig } from '../../models/invitation';
import { InvitationService } from '../../services/invitation';
import { DocumentComponent } from '../document/document';

@Component({
  selector: 'app-creator',
  imports: [CommonModule, FormsModule, DocumentComponent],
  templateUrl: './creator.html'
})
export class CreatorComponent implements OnInit {
  private readonly invitationService = inject(InvitationService);
  private readonly router = inject(Router);
  private readonly route = inject(ActivatedRoute);

  // Loading and Save states
  readonly saving = signal(false);
  readonly savedId = signal<string | null>(null);

  isEditMode = false;
  invitationId: string | null = null;

  // Form Model
  invitation: Invitation = {
    type: 'remito',
    docNumber: this.generateRandomDocNumber(),
    emisor: {
      name: 'Alejandro Magno',
      company: 'POTENCIAPP',
      cuit: '30-71452968-3',
      address: 'Piso 12, Av. del Libertador 4242, CABA',
      email: 'colazurdaalejo@potenciapp.com'
    },
    receptor: {
      name: 'Equipo de Desarrollo',
      company: 'Sistemas Internos',
      cuit: '30-99999999-9',
      address: 'Oficinas 3er Piso, Bloque B',
      email: 'devteam@empresa.com'
    },
    eventDate: this.getDefaultEventDate(),
    items: [],
    comments: 'Al firmar este remito, el receptor se compromete a asistir al evento y a traer excelente actitud. El emisor no se responsabiliza por sobredosis de cafeína o medialunas.',
    theme: 'classic',
    layout: {
      pageSize: 'A4',
      widthMm: 210,
      heightMm: 297,
      marginTop: 10,
      marginBottom: 10,
      marginLeft: 10,
      marginRight: 10,
      colQtyWidth: 15,
      colDescWidth: 50,
      colPriceWidth: 18,
      colTotalWidth: 17,
      fontSizeScale: 1.0
    }
  };

  // Preset models for easy setup
  selectedPreset = 'coffee';

  constructor() {}

  ngOnInit() {
    this.route.paramMap.subscribe(params => {
      const id = params.get('id');
      if (id) {
        this.isEditMode = true;
        this.invitationId = id;
        this.loadInvitation(id);
      } else {
        this.isEditMode = false;
        this.invitationId = null;
        this.applyPreset('coffee');
      }
    });
  }

  loadInvitation(id: string) {
    this.invitationService.getInvitation(id).subscribe({
      next: (inv) => {
        let localDate = '';
        if (inv.eventDate) {
          try {
            const d = new Date(inv.eventDate);
            const tzOffset = d.getTimezoneOffset() * 60000;
            localDate = new Date(d.getTime() - tzOffset).toISOString().slice(0, 16);
          } catch {
            localDate = inv.eventDate;
          }
        }
        this.invitation = {
          ...inv,
          eventDate: localDate
        };
        this.selectedPreset = 'custom';
      },
      error: (err) => {
        console.error('Error loading invitation for editing:', err);
        alert('No se pudo cargar la invitación para editar.');
        this.router.navigate(['/']);
      }
    });
  }

  generateRandomDocNumber(): string {
    const prefix = '0001';
    const num = Math.floor(100000 + Math.random() * 900000);
    return `${prefix}-${num}`;
  }

  getDefaultEventDate(): string {
    const d = new Date();
    d.setDate(d.getDate() + 1); // tomorrow
    d.setHours(16, 0, 0, 0); // 16:00
    // Format to local date picker format 'YYYY-MM-DDThh:mm'
    const tzOffset = d.getTimezoneOffset() * 60000;
    const localISOTime = new Date(d.getTime() - tzOffset).toISOString().slice(0, 16);
    return localISOTime;
  }

  applyPreset(presetType: string) {
    this.selectedPreset = presetType;
    if (presetType === 'coffee') {
      this.invitation.theme = 'coffee';
      this.invitation.comments = 'Al firmar este remito, el receptor se compromete a levantarse de su silla, conversar sin hablar de bugs y consumir un mínimo de 2 medialunas.';
      this.invitation.items = [
        { qty: 1, description: 'Café de especialidad (Caliente y Espumoso)', unitPrice: 0, total: 0 },
        { qty: 2, description: 'Medialunas de manteca calentitas (Relleno opcional)', unitPrice: 0, total: 0 },
        { qty: 15, description: 'Minutos de desconexión mental y risas con colegas', unitPrice: 0, total: 0 }
      ];
    } else if (presetType === 'after') {
      this.invitation.theme = 'blue';
      this.invitation.comments = 'Condición de entrega: Los productos deben consumirse helados. Queda terminantemente prohibido hablar de pull requests o estimaciones de Jira.';
      this.invitation.items = [
        { qty: 2, description: 'Pintas de Cerveza Artesanal (IPA/Golden bien heladas)', unitPrice: 0, total: 0 },
        { qty: 1, description: 'Porción individual de Papas Rústicas con Cheddar y Panceta', unitPrice: 0, total: 0 },
        { qty: 1, description: 'Entorno de socialización y networking descontracturado', unitPrice: 0, total: 0 }
      ];
    } else if (presetType === 'asado') {
      this.invitation.theme = 'classic';
      this.invitation.comments = 'Al firmar este remito, el receptor acepta traer apetito. Se aceptan sugerencias para el playlist del quincho.';
      this.invitation.items = [
        { qty: 1, description: 'Choripán con chimichurri casero de bienvenida', unitPrice: 0, total: 0 },
        { qty: 1, description: 'Corte de Asado / Vacío a punto de cruz con ensalada criolla', unitPrice: 0, total: 0 },
        { qty: 1, description: 'Bebida a elección helada (Vino/Gaseosa/Agua)', unitPrice: 0, total: 0 }
      ];
    } else {
      // Custom - empty lists
      this.invitation.theme = 'classic';
      this.invitation.items = [
        { qty: 1, description: 'Corte/Invitación Personalizada', unitPrice: 0, total: 0 }
      ];
    }
  }

  // List editing functions
  addItem() {
    this.invitation.items.push({
      qty: 1,
      description: '',
      unitPrice: 0,
      total: 0
    });
  }

  removeItem(index: number) {
    if (this.invitation.items.length > 1) {
      this.invitation.items.splice(index, 1);
    }
  }

  // Adjust table layout columns proportionally
  adjustColumns(columnChanged: string) {
    const l = this.invitation.layout;
    
    // Ensure sum is always 100.
    // In our slider model, the user modifies Qty, Desc, and Price. Total takes the remainder.
    // Limit them to safe ranges.
    if (l.colQtyWidth < 10) l.colQtyWidth = 10;
    if (l.colPriceWidth < 10) l.colPriceWidth = 10;
    if (l.colDescWidth < 30) l.colDescWidth = 30;

    const remaining = 100 - (l.colQtyWidth + l.colDescWidth + l.colPriceWidth);
    if (remaining >= 10) {
      l.colTotalWidth = remaining;
    } else {
      // Scale down Desc to preserve Total minimum
      l.colTotalWidth = 10;
      l.colDescWidth = 100 - (l.colQtyWidth + l.colPriceWidth + l.colTotalWidth);
    }
  }

  onPageSizeChange() {
    const l = this.invitation.layout;
    if (l.pageSize === 'A4') {
      l.widthMm = 210;
      l.heightMm = 297;
    } else if (l.pageSize === 'A5') {
      l.widthMm = 148;
      l.heightMm = 210;
    } else if (l.pageSize === 'Letter') {
      l.widthMm = 216;
      l.heightMm = 279;
    }
  }

  readonly isDraggingLogo = signal(false);

  onDragOverLogo(event: DragEvent) {
    event.preventDefault();
    event.stopPropagation();
    this.isDraggingLogo.set(true);
  }

  onDragLeaveLogo(event: DragEvent) {
    event.preventDefault();
    event.stopPropagation();
    this.isDraggingLogo.set(false);
  }

  onDropLogo(event: DragEvent) {
    event.preventDefault();
    event.stopPropagation();
    this.isDraggingLogo.set(false);

    if (event.dataTransfer && event.dataTransfer.files && event.dataTransfer.files[0]) {
      const file = event.dataTransfer.files[0];
      if (file.type.startsWith('image/')) {
        const reader = new FileReader();
        reader.onload = (e) => {
          this.invitation.emisor.logo = reader.result as string;
        };
        reader.readAsDataURL(file);
      } else {
        alert('Por favor, sube únicamente archivos de tipo imagen.');
      }
    }
  }

  onLogoUpload(event: Event) {
    const input = event.target as HTMLInputElement;
    if (input.files && input.files[0]) {
      const file = input.files[0];
      const reader = new FileReader();
      reader.onload = (e) => {
        this.invitation.emisor.logo = reader.result as string;
      };
      reader.readAsDataURL(file);
    }
  }

  removeLogo() {
    this.invitation.emisor.logo = '';
  }

  saveInvitation() {
    this.saving.set(true);
    
    // Convert local selector date to ISO UTC Date for backend
    const invToSave = { ...this.invitation };
    try {
      invToSave.eventDate = new Date(this.invitation.eventDate).toISOString();
    } catch {
      // Use raw input if conversion fails
    }

    if (this.isEditMode && this.invitationId) {
      this.invitationService.updateInvitation(this.invitationId, invToSave).subscribe({
        next: (res) => {
          this.saving.set(false);
          this.savedId.set(res.id || this.invitationId);
        },
        error: (err) => {
          this.saving.set(false);
          alert('Ocurrió un error al actualizar la invitación. Revisa la conexión con el servidor Go.');
          console.error(err);
        }
      });
    } else {
      this.invitationService.createInvitation(invToSave).subscribe({
        next: (res) => {
          this.saving.set(false);
          this.savedId.set(res.id || null);
        },
        error: (err) => {
          this.saving.set(false);
          alert('Ocurrió un error al guardar la invitación. Revisa la conexión con el servidor Go.');
          console.error(err);
        }
      });
    }
  }

  getSharedLink(): string {
    const base = window.location.origin;
    return `${base}/invitation/${this.savedId()}`;
  }

  copyLinkToClipboard() {
    const link = this.getSharedLink();
    navigator.clipboard.writeText(link).then(() => {
      alert('¡Enlace de invitación copiado al portapapeles!');
    }).catch(err => {
      console.error('Error copying text: ', err);
    });
  }

  downloadPDF() {
    if (this.savedId()) {
      window.open(this.invitationService.getPDFUrl(this.savedId()!), '_blank');
    }
  }

  goToDashboard() {
    this.router.navigate(['/']);
  }
}
