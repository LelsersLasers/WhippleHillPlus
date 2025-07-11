<script>
    import TimezonePicker from 'svelte-timezone-picker';
    import Modal from "./Modal.svelte";
    import { fly } from "svelte/transition"; 
    import { onDestroy } from "svelte";

    export let data;
    export let api;

    const days = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];

    let page = "assignments";

    let assignments = [];
    let classes = [];
    let semesters = [];
    let user_name = "Loading...";
    let ics_link = "";

    let semester = -1;
    let semesterValue = -1;

    let shownAssignments = [];
    let shownClasses = [];
    let classFilter = [];

    let unique = {};

    let timezone = "";

    let isSmallScreen = window.matchMedia("(max-width: 600px)").matches;

    const mediaQuery = window.matchMedia("(max-width: 600px)");
	const update = () => {
		isSmallScreen = mediaQuery.matches;
	};

	mediaQuery.addEventListener("change", update);

	onDestroy(() => {
		mediaQuery.removeEventListener("change", update);
	});

    processMainData(data);

    function processMainData(f) {
        f
            .then((res) => res.json())
            .then((data) => {
                console.log(data);

                if (data["error"] && data["error"] != "") {
                    window.location.href = "/login";
                    return;
                }

                const oldAssignments = assignments;
                const oldClasses = classes;

                assignments = data["assignments"];
                classes = data["classes"];
                semesters = data["semesters"];
                user_name = data["user_name"];
                if (data["ics_link"]) {
                    // ics_link = `${api}/ics/${data["ics_link"]}.ics`;
                    ics_link = data["ics_link"];
                }
                if (data["timezone"]) {
                    timezone = data["timezone"];
                } else {
                    timezone = Intl.DateTimeFormat().resolvedOptions().timeZone;
                    updateTimezone({ detail: { timezone: timezone } });
                }

                if (semester == -1) {
                    semesters = semesters.sort((a, b) => b.sort_order - a.sort_order);
                    if (semesters[0]) {
                        semester = semesters[0].id;
                        semesterValue = semester;
                    }
                }

                assignments.forEach((a) => {
                    if (unique[a.id] === undefined) {
                        unique[a.id] = 0;
                    } else {
                        // if assigment has changed proc the key to replay the animation
                        convertToLocalTime(a);
                        const oldAssignment = oldAssignments.find((o) => o.id === a.id);
                        if (!assigmentsAreEqual(a, oldAssignment)) {
                            unique[a.id] += 1;
                        }
                    }
                });

                classes.forEach((c) => {
                    if (!oldClasses.find((o) => o.id === c.id)) {
                        classFilter = [...classFilter, c.id];
                    }
                });
            });
    }

    function assigmentsAreEqual(a, b) {
        return a.id === b.id
            && a.name === b.name
            && a.description === b.description
            && a.assigned_date === b.assigned_date
            && a.due_date === b.due_date
            && a.due_time === b.due_time
            && a.status === b.status
            && a.type === b.type
            && a.class_id === b.class_id;
    }

    function updateSemesterValue(e) {
        semester = e.target.value;
    }

    function convertToLocalTime(a) {
        if (a.due_date.endsWith("Z")) {
            a.due_date = a.due_date.slice(0, -1);
        }
        if (a.assigned_date.endsWith("Z")) {
            a.assigned_date = a.assigned_date.slice(0, -1);
        }
    }

    function convertTimeTo12Hours(time) {
        const [hours, minutes] = time.split(":");
        const hour = parseInt(hours);
        if (hour === 0) {
            return `12:${minutes}am`;
        } else if (hour < 12) {
            return `${hour}:${minutes}am`;
        } else if (hour === 12) {
            return `12:${minutes}pm`;
        } else {
            return `${hour - 12}:${minutes}pm`;
        }
    }

    function sortClasses(a, b) {
        // Sort "other" to the bottom
        // Otherwise alphabetically
        if (a.name === b.name) return 0;
        if (a.name.toLowerCase() === "other") return 1;
        if (b.name.toLowerCase() === "other") return -1;
        return a.name.localeCompare(b.name);
    }

    $: {
        semesters = semesters.sort((a, b) => b.sort_order - a.sort_order);
    }

    $: {
        classes = classes.sort(sortClasses);
    }
    $: {
        shownClasses = classes.filter((c) => c.semester_id == semester);
    }
    $: {
        classFilter = classFilter.sort((a, b) => classFromId(a).name.localeCompare(classFromId(b).name));
    }

    $: {
        function sortAssignments(a, b) {
            // First by date
            if (a.due_date < b.due_date) return -1;
            if (a.due_date > b.due_date) return 1;
            
            // Then by time
            if (a.due_time != "" && b.due_time == "") return -1;
            if (a.due_time == "" && b.due_time != "") return 1;
            if (a.due_time < b.due_time) return -1;
            if (a.due_time > b.due_time) return 1;

            // Then by status
            const statusSeight = {
                "Completed":   1,
                "In Progress": 0,
                "Not Started": 0,
            };
            if (statusSeight[a.status] < statusSeight[b.status]) return -1;
            if (statusSeight[a.status] > statusSeight[b.status]) return 1;

            // Then by type
            const typeWeight = {
                "Other":    5,
                "Homework": 4,
                "Project":  3,
                "Paper":    2,
                "Quiz":     1,
                "Test":     0,
            };
            if (typeWeight[a.type] < typeWeight[b.type]) return -1;
            if (typeWeight[a.type] > typeWeight[b.type]) return 1;

            // Then by class
            const classA = classes.find((c) => c.id === a.class_id);
            const classB = classes.find((c) => c.id === b.class_id);
            const classSort = sortClasses(classA, classB);
            if (classSort !== 0) return classSort;

            // lastly by name
            return a.name.localeCompare(b.name);
        }
        
        assignments.forEach(convertToLocalTime);
        assignments.sort(sortAssignments);
        assignments = assignments;
    }

    function localDate() {
        const date = new Date();
        const timeOffset = date.getTimezoneOffset();
        const hoursOffset = Math.floor(timeOffset / 60);
        const minutesOffset = timeOffset % 60;
        date.setHours(date.getHours() - hoursOffset);
        date.setMinutes(date.getMinutes() - minutesOffset);
        return date;
    }

    let showICSModal = false;
    let generateICSLinkButton = true;
    
    let showCreateSemesterModal = false;
    let createSemesterModalButton = true;

    let showAllSemestersModal = false;

    let showUpdateSemesterModal = false;
    let updateSemesterModalName = "";
    let updateSemesterModalSortOrder = "1";
    let updateSemesterModalId = "";
    let updateSemesterModalButton = true;

    let deleteSemesterModalButton = true;

    let showCreateClassModal = false;
    let createClassModalButton = true;

    let showUpdateClassModal = false;
    let updateClassModalName = "";
    let updateClassModalSemesterID = -1;
    let updateClassModalId = "";
    let updateClassModalButton = true;

    let showDeleteClassModal = false;
    let deleteClassModalName = "";
    let deleteClassModalId = "";
    let deleteClassModalTimer = 15;
    let deleteClassModalTimerInterval = null;
    let deleteClassModalButton = true;


    let showCreateAssignmentModal = false;
    let createAssignmentModalButton = true;

    let showUpdateAssignmentModal = false;
    let updateAssignmentModalName = "";
    let updateAssignmentModalDescription = "";
    let updateAssignmentModalAssignedDate = formatDateObj(localDate());
    let updateAssignmentModalDueDate = "";
    let updateAssignmentModalDueTime = "";
    let updateAssignmentModalStatus = "";
    let updateAssignmentModalType = "";
    let updateAssignmentModalClassId = "";
    let updateAssignmentModalId = "";
    let updateAssignmentModalButton = true;
    let updateAssignmentModalDeleteButton = true;

    let showAssignmentDetailsModal = false;
    let assignmentDetailsModalName = "";
    let assignmentDetailsModalId = "";
    let assignmentDetailsModalDescription = "";
    let assignmentDetailsModalAssignedDate = "";
    let assignmentDetailsModalDueDate = "";
    let assignmentDetailsModalDueTime = "";
    let assignmentDetailsModalStatus = "";
    let assignmentDetailsModalType = "";
    let assignmentDetailsModalClassId = "";

    let showDateFilterModal = false;
    let dateFilter = "future";
    let rangeIncludesAssigned = false;

    let showClassFilterModal = false;

    let showStatusFilter = false;
    let statusFilter = ["Not Started", "In Progress", "Completed"];

    let showTypeFilter = false;
    let typeFilter = ["Homework", "Quiz", "Test", "Project", "Paper", "Other"];

    let [today, pastSunday, nextSunday] = createDates();

    let dateStart = formatDateObj(pastSunday);
    let dateEnd = formatDateObj(nextSunday);

    function missingCheck(a) {
        // due date is in the past and status is not completed
        if (a.status == "Completed") return false;

        const dueDate = new Date(a.due_date);
        if (a.due_time != "") {
            const todayWithTime = new Date();

            const dueTime = a.due_time.split(":");
            dueDate.setHours(dueTime[0]);
            dueDate.setMinutes(dueTime[1]);

            return dueDate < todayWithTime;
        } else {
            return dueDate < today;
        }
    }

    function assignmentToColor(a) {
        const missing = missingCheck(a);
        if (missing) return "#bf616a";
        const status_color = {
            "Not Started": "#b48ead",
            "In Progress": "#ebcb8b",
            "Completed": "#a3be8c",
        };
        return status_color[a.status];
    }

    $: {
        function semesterCheck(a) {
            const c = classes.find((c) => c.id === a.class_id);
            return c.semester_id == semester;
        }

        function classFilterCheck(a) {
            return classFilter.includes(a.class_id);
        }
        function statusFilterCheck(a) {
            return statusFilter.includes(a.status);
        }
        function typeFilterCheck(a) {
            return typeFilter.includes(a.type);
        }
        function datesOverlap(start1, end1, start2, end2) {
            return start1 <= end2 && end1 >= start2;
        }

        if (dateFilter == "all") {
            shownAssignments = assignments.filter((a) => semesterCheck(a));
        } else if (dateFilter == "week") {
            let dateWeekStart = formatDateObj(pastSunday);
            let dateWeekEnd = formatDateObj(nextSunday);

            shownAssignments = assignments.filter((a) => {
                const dueDate = new Date(a.due_date);
                const assignedDate = new Date(a.assigned_date);
                if (rangeIncludesAssigned) {
                    const overlaps = datesOverlap(start1, end1, assignedDate, dueDate);
                    return (semesterCheck(a) && classFilterCheck(a) && statusFilterCheck(a) && typeFilterCheck(a) && overlaps) || missingCheck(a);
                } else {
                    const dueDateInRange = dueDate >= new Date(dateWeekStart) && dueDate <= new Date(dateWeekEnd);
                    return (semesterCheck(a) && classFilterCheck(a) && statusFilterCheck(a) && typeFilterCheck(a) && dueDateInRange) || missingCheck(a);
                }
            });
        } else if (dateFilter == "future") {
            shownAssignments = assignments.filter((a) => {
                const dueDate = new Date(a.due_date);
                return (semesterCheck(a) && classFilterCheck(a) && statusFilterCheck(a) && typeFilterCheck(a) && dueDate >= today) || missingCheck(a);
            });
        } else if (dateFilter == "range") {
            shownAssignments = assignments.filter((a) => {
                const dueDate = new Date(a.due_date);
                const assignedDate = new Date(a.assigned_date);
                if (rangeIncludesAssigned) {
                    const overlaps = datesOverlap(new Date(dateStart), new Date(dateEnd), assignedDate, dueDate);
                    return (semesterCheck(a) && classFilterCheck(a) && statusFilterCheck(a) && typeFilterCheck(a) && overlaps) || missingCheck(a);
                } else {
                    const dueDateInRange = dueDate >= new Date(dateStart) && dueDate <= new Date(dateEnd);
                    return (semesterCheck(a) && classFilterCheck(a) && statusFilterCheck(a) && typeFilterCheck(a) && dueDateInRange) || missingCheck(a);
                }
            });
        }
    }

    function createDates() {
        let todayDate = new Date();
        todayDate.setHours(0, 0, 0, 0);

        let pastSundayDate = new Date(todayDate);
        pastSundayDate.setDate(todayDate.getDate() - todayDate.getDay());

        let nextSundayDate = new Date(todayDate);
        if (7 - todayDate.getDay() < 2) {
            nextSundayDate.setDate(todayDate.getDate() + (21 - todayDate.getDay()));
        } else {
            nextSundayDate.setDate(todayDate.getDate() + (14 - todayDate.getDay()));
        }

        return [todayDate, pastSundayDate, nextSundayDate];
    }


    function formatDateObj(date) {
        // Date to string in the format "yyyy-MM-dd"
        let year = date.getFullYear();
        let month = String(date.getMonth() + 1).padStart(2, '0');
        let day = String(date.getDate()).padStart(2, '0');
        
        return `${year}-${month}-${day}`;
    }
    
    function formatDateString(str) {
        // "yyyy-MM-dd" -> "m/d/yy"
        let [year, month, day] = str.split("-");
        return `${parseInt(month)}/${parseInt(day)}/${year.slice(2)}`;
    }


    function formDataWithoutReload(e) {
        e.preventDefault();

        const form = e.target;
        const formData = new FormData(form);
        const data = Object.fromEntries(formData);
        
        return data;
    }

    function classFromId(id) {
        return classes.find((c) => c.id === id);
    }

    function deleteClassButton(id) {
        const data = {
            'id': id,
        }
        deleteClassModalButton = false;

        fetch(`${api}/delete_class`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(data),
        })
            .then((res) => {
                classes = classes.filter((c) => c.id !== id);
                assignments = assignments.filter((a) => a.class_id !== id);
                classFilter = classFilter.filter((c) => c !== id);

                deleteClassModalName = "";
                deleteClassModalId = "";
                deleteClassModalTimer = 0;
                showDeleteClassModal = false;
                deleteClassModalButton = true;
            })
    }
    function deleteModalButton(id) {
        const c = classes.find((c) => c.id === id);
        deleteClassModalName = c.name;
        deleteClassModalId = c.id;
        deleteClassModalTimer = 15;
        if (deleteClassModalTimerInterval) clearInterval(deleteClassModalTimerInterval);
        deleteClassModalTimerInterval = setInterval(() => {
            deleteClassModalTimer -= 1;
            if (deleteClassModalTimer <= 0) {
                clearInterval(deleteClassModalTimerInterval);
            }
        }, 1000);
        showDeleteClassModal = true;
    }
    function updateClassButton(id) {
        const c = classes.find((c) => c.id === id);
        updateClassModalName = c.name;
        updateClassModalSemesterID = c.semester_id;
        updateClassModalId = c.id;
        showUpdateClassModal = true;
    }
    function createClass(e) {
        const data = formDataWithoutReload(e);
        createClassModalButton = false;
        
        fetch(`${api}/create_class`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(data),
        })
            .then((res) =>res.json())
            .then((data) => {
                classes = [...classes, data];
                classFilter = [...classFilter, data.id];
                document.getElementById("createClassModalName").value = "";
                // document.getElementById("createClassModalSemesterID").value = semester;
                showCreateClassModal = false;
                createClassModalButton = true;
            })
    }
    function updateClass(e) {
        const data = formDataWithoutReload(e);
        updateClassModalButton = false;

        fetch(`${api}/update_class`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(data),
        })
            .then((res) => res.json())
            .then((data) => {
                classes = classes.map((c) => {
                    if (c.id === data.id) return data;
                    else                  return c;
                });
                assignments = assignments;
                updateClassModalName = "";
                updateClassModalSemesterID = -1;
                updateClassModalId = "";
                showUpdateClassModal = false;
                updateClassModalButton = true;
            })
    }

    function createSemester(e) {
        const data = formDataWithoutReload(e);
        createSemesterModalButton = false;

        fetch(`${api}/create_semester`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(data),
        })
            .then((res) =>res.json())
            .then((data) => {
                semesters = data;
                if (semesters.length == 1) {
                    semester = semesters[0].id;
                }
                
                document.getElementById("createSemesterModalName").value = "";
                document.getElementById("createSemesterModalSortOrder").value = "1";
                showCreateSemesterModal = false;
                createSemesterModalButton = true;
            })
    }
    function updateSemester(e) {
        const data = formDataWithoutReload(e);
        updateSemesterModalButton = false;

        console.log(data);

        fetch(`${api}/update_semester`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(data),
        })
            .then((res) => res.json())
            .then((data) => {
                semesters = data;

                updateSemesterModalName = "";
                updateSemesterModalSortOrder = "1";
                showUpdateSemesterModal = false;
                updateSemesterModalButton = true;
            })
    }
    function updateSemesterButton(id) {
        const s = semesters.find((s) => s.id === id);
        updateSemesterModalName = s.name;
        updateSemesterModalSortOrder = s.sort_order;
        updateSemesterModalId = s.id;
        showUpdateSemesterModal = true;
    }
    function deleteModalButtonSemester(id) {
        if (classes.find((c) => c.semester_id == id)) {
            alert("Cannot delete a semester with classes in it.");
            return;
        }

        if (semesters.length == 1) {
            alert("Cannot delete the last semester.");
            return;
        }

        deleteSemesterModalButton = false;

        const data = { 'id': id };
        fetch(`${api}/delete_semester`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(data),
        })
            .then((res) => res.json())
            .then((data) => {
                semesters = data;
            
                if (semester == id && semesters[0]) {
                    semester = semesters[0].id;
                    semesterValue = semester;
                }

                deleteSemesterModalButton = true;
            });
    }

    function deleteAssignmentButton(id) {
        const data = {
            'id': id,
        }
        updateAssignmentModalDeleteButton = false;

        fetch(`${api}/delete_assignment`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(data),
        })
            .then((res) => {
                assignments = assignments.filter((a) => a.id !== id);
                unique[id] += 1;
            
                updateAssignmentModalName = "";
                updateAssignmentModalDescription = "";
                updateAssignmentModalAssignedDate = formatDateObj(localDate());
                updateAssignmentModalDueDate = "";
                updateAssignmentModalDueTime = "";
                updateAssignmentModalStatus = "";
                updateAssignmentModalType = "";
                updateAssignmentModalClassId = "";
                showUpdateAssignmentModal = false;
                updateAssignmentModalDeleteButton = true;
            })
    }
    function updateAssignmentButton(id) {
        const a = assignments.find((a) => a.id === id);
        updateAssignmentModalName = a.name;
        updateAssignmentModalDescription = a.description;
        updateAssignmentModalAssignedDate = a.assigned_date.slice(0, "yyyy-MM-dd".length);
        updateAssignmentModalDueDate = a.due_date.slice(0, "yyyy-MM-dd".length);
        updateAssignmentModalDueTime = a.due_time;
        updateAssignmentModalStatus = a.status;
        updateAssignmentModalType = a.type;
        updateAssignmentModalClassId = a.class_id;
        updateAssignmentModalId = a.id;
        showUpdateAssignmentModal = true;
    }
    function assignmentDetailsButton(id) {
        const a = assignments.find((a) => a.id === id);
        assignmentDetailsModalName = a.name;
        assignmentDetailsModalId = a.id;
        assignmentDetailsModalDescription = a.description;
        assignmentDetailsModalAssignedDate = a.assigned_date;
        assignmentDetailsModalDueDate = a.due_date;
        assignmentDetailsModalDueTime = a.due_time;
        assignmentDetailsModalStatus = a.status;
        assignmentDetailsModalType = a.type;
        assignmentDetailsModalClassId = a.class_id;
        showAssignmentDetailsModal = true;
    }
    function createAssignment(e) {
        const data = formDataWithoutReload(e);
        createAssignmentModalButton = false;

        fetch(`${api}/create_assignment`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(data),
        })
            .then((res) => res.json())
            .then((data) => {
                assignments = [...assignments, data];
                unique[data.id] += 1;

                document.getElementById("createAssignmentModalName").value = "";
                document.getElementById("createAssignmentModalDescription").value = "";
                document.getElementById("createAssignmentModalAssignedDate").valueAsDate = localDate();
                document.getElementById("createAssignmentModalDueDate").value = "";
                document.getElementById("createAssignmentModalDueTime").value = "";
                document.getElementById("createAssignmentModalStatus").value = "Not Started";
                document.getElementById("createAssignmentModalType").value = "Homework";
                // document.getElementById("createAssignmentModalClassId").value = "";
                showCreateAssignmentModal = false;
                createAssignmentModalButton = true;
            })
    };
    function statusAssignment(e, id) {
        const data = {
            'id': id,
            'status': e.target.value,
        }

        fetch(`${api}/status_assignment`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(data),
        })
            .then((res) => res.json())
            .then((data) => {
                assignments = assignments.map((a) => {
                    if (a.id === data.id) return data;
                    else                  return a;
                });
                unique[data.id] += 1;
            })
    }
    function updateAssignment(e) {
        const data = formDataWithoutReload(e);
        updateAssignmentModalButton = false;

        fetch(`${api}/update_assignment`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(data),
        })
            .then((res) => res.json())
            .then((data) => {
                assignments = assignments.map((a) => {
                    if (a.id === data.id) return data;
                    else                  return a;
                });
                unique[data.id] += 1;

                updateAssignmentModalName = "";
                updateAssignmentModalDescription = "";
                updateAssignmentModalAssignedDate = formatDateObj(localDate());
                updateAssignmentModalDueDate = "";
                updateAssignmentModalDueTime = "";
                updateAssignmentModalStatus = "";
                updateAssignmentModalType = "";
                updateAssignmentModalClassId = "";
                showUpdateAssignmentModal = false;
                updateAssignmentModalButton = true;

                assignmentDetailsModalName = data.name;
                assignmentDetailsModalDescription = data.description;
                assignmentDetailsModalAssignedDate = data.assigned_date;
                assignmentDetailsModalDueDate = data.due_date;
                assignmentDetailsModalDueTime = data.due_time;
                assignmentDetailsModalStatus = data.status;
                assignmentDetailsModalType = data.type;
                assignmentDetailsModalClassId = data.class_id;
            })
    }


    function generateICSLink() {
        if (ics_link != "") {
            if (!window.confirm("Are you sure you want to generate a new ICS link? This will invalidate the old ones. You do not need to regenerate the link if you just want to update the timezone.")) {
                return;
            }
        }

        generateICSLinkButton = false;

        fetch(`${api}/ics/generate`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
        })
            .then((res) => res.json())
            .then((data) => {
                // ics_link = `${api}/ics/${data["ics_link"]}.ics`;
                ics_link = data["ics_link"];
                generateICSLinkButton = true;
            })
    }

    function updateTimezone(ev) {
        generateICSLinkButton = false;
        const data = {
            'timezone': ev.detail.timezone,
        }

        fetch(`${api}/ics/update_timezone`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(data),
        })
            .then((res) => res.json())
            .then((data) => {
                timezone = data["timezone"];
                generateICSLinkButton = true;
            })
    }

    function dateToDayOfWeek(date) {
        const dayIndex = new Date(date).getDay();
        return days[dayIndex];
    }


    addEventListener("DOMContentLoaded", () => {
        document.getElementById("createClass").addEventListener("submit", createClass);
        document.getElementById("updateClass").addEventListener("submit", updateClass);

        document.getElementById("createSemester").addEventListener("submit", createSemester);
        document.getElementById("updateSemester").addEventListener("submit", updateSemester);

        document.getElementById("createAssignmentModalAssignedDate").valueAsDate = localDate();
        document.getElementById("createAssignment").addEventListener("submit", createAssignment);
        document.getElementById("updateAssignment").addEventListener("submit", updateAssignment);

        document.addEventListener("visibilitychange", () => {
            if (!document.hidden) {
                processMainData(fetch(`${api}/home_data`));
                [today, pastSunday, nextSunday] = createDates();
            }
            document.getElementById("createAssignmentModalAssignedDate").valueAsDate = localDate();
        });
    });
</script>


<style>

#logout {
    float: right;
}

#holder {
    width: 100%;
    margin: 0 auto;
}

@media (min-width: 800px) {
    #holder {
        width: 80%;
    }
}


.description {
    white-space: pre-line;
}

table {
    /* width: 80%;
    margin: 0 auto; */
    width: 100%;
    border-collapse: collapse;
}

th {
    background-color: #afafaf;
    text-align: left;
    padding-left: 0;
    border-bottom: 2px black solid;
}

tr:nth-child(even) {
    background-color: #f4f4f4;
}

.breakWord {
    word-break: break-word;
}


.zeroWidth {
    width: 0;
}
.classWidth {
    width: 15vw;
}

.untightPadding {
    padding-right: 15px;
}

.pointer {
    cursor: pointer;
}

#createSemesterButton {
    margin-top: 12px;
}


input[type="text"] {
    width: 100%;
}
textarea {
    width: 100%;
    height: 5em;
}


select:active {
    background-color: revert !important;
}

button:disabled {
    animation: bob 0.5s infinite ease-in-out;
}

button[type="submit"] {
    font-weight: 600;
}

@keyframes bob {
    0% {
        transform: translateY(0);
    }
    50% {
        transform: translateY(-5px);
    }
    100% {
        transform: translateY(0);
    }
}


.calenderLabel {
    text-align: center;
    margin-bottom: 4px;
}
code {
    background-color: #d4d4d4;
    padding: 0.5em;
    border-radius: 5px;
    width: 100%;
    text-align: center;
    display: block;
}

.mobile-row {
    display: flex;
    flex-direction: row;
    justify-content: space-between;
    align-items: center;
    padding: 0.5em 0;
    word-break: break-word;
    /* border-bottom: 1px solid #9893a5; */
}

.mobile-primary {
    flex-grow: 1;
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    padding-left: 0.25em;
}

.mobile-title {
    font-size: 1em;
    margin: 0;
}

.mobile-link {
    margin-left: 0.5em;
    text-decoration: none;
}

.mobile-edit {
    font-size: 0.7em;
    color: #666;
    cursor: pointer;
}

.mobile-details {
    display: flex;
    flex-direction: row;
    gap: 0.5em;
    font-size: 0.75em;
    color: #666;
    align-items: flex-start;
    flex-wrap: wrap;
}

.mobile-due {
    color: #000;
}

.mobile-status {
    font-size: 1em;
    word-break: normal;
    margin-left: 0.5em;
    text-align: right;
    margin-right: 0.25em;
    flex: 0 0 auto;
}

</style>



<div id="holder">
<h1>Welcome, {user_name}!</h1>

<select id="semesterSelector" name="semesterSelector" on:change={updateSemesterValue} bind:value={semesterValue}>
    {#each semesters as s (s.id)}
        <option value={s.id}>{s.name}</option>
    {/each}
</select>

<button id="logout" type="button" on:click={() => window.location.href = "/logout_user"}>Logout</button>

{#if page == "classes"}
    <h2>Your Classes</h2>
    <button type="button" on:click={() => page = "assignments"}>View Assignments</button>
    <button type="button" on:click={() => { showCreateClassModal = true; document.getElementById("createClassModalSemesterID").value = semester; }}>Create Class</button>
    <button type="button" on:click={() => showAllSemestersModal = true}>Semesters</button>
    <button type="button" on:click={() => showICSModal = true}>Calendar Integration</button>

    <table>
        <tr>
            <th>Name</th>
            <th class="zeroWidth"></th>
            <th class="zeroWidth"></th>
        </tr>
        {#each shownClasses as c (c.id)}
            <tr
                in:fly|global={{ duration: 300, x: -200 }}
            >
                <td class="breakWord">{c.name}</td>
                <td class="zeroWidth">
                    <button type="button" on:click={() => updateClassButton(c.id)}>Edit</button>
                </td>
                <td class="zeroWidth">
                    <button type="button" on:click={() => deleteModalButton(c.id)}>Delete</button>
                </td>
            </tr>
        {/each}
    </table>
{:else if page == "assignments"}
    <h2>Your Assignments</h2>
    <button type="button" on:click={() => page = "classes"}>View Classes</button>
    <button type="button" on:click={() => showCreateAssignmentModal = true}>Create Assignment</button>
    <button type="button" on:click={() => showDateFilterModal = true}>Date Filter</button>
    <button type="button" on:click={() => showClassFilterModal = true}>Class Filter</button>
    <button type="button" on:click={() => showStatusFilter = true}>Status Filter</button>
    <button type="button" on:click={() => showTypeFilter = true}>Type Filter</button>

    {#if isSmallScreen}
        {#each shownAssignments as a, i (a.id)}
            <div
                class="mobile-row"
                style="{shownAssignments[i + 1] && a.due_date != shownAssignments[i + 1].due_date ? 'border-bottom: 1px solid black;' : ''}"
            >
                <div class="mobile-primary">
                    <p class="mobile-title">
                        <!-- svelte-ignore a11y-click-events-have-key-events -->
                        <!-- svelte-ignore a11y-missing-attribute -->
                        <a on:click={() => assignmentDetailsButton(a.id)}>{a.name}</a>
                        <!-- svelte-ignore a11y-click-events-have-key-events -->
                        <!-- svelte-ignore a11y-missing-attribute -->
                        <a class="mobile-link" on:click={() => updateAssignmentButton(a.id)}>
                            <span class="mobile-edit">(edit)</span>
                        </a>
                    </p>
                    <div class="mobile-details">
                        <div>
                            {a.type}: {classFromId(a.class_id).name}
                        </div>
                    </div>
                    <div class="mobile-details">
                        {#if a.due_time != ""}
                            Due: <span class="mobile-due">{formatDateString(a.due_date)}</span> ({dateToDayOfWeek(a.due_date)}) at {convertTimeTo12Hours(a.due_time)}
                        {:else}
                            Due: <span class="mobile-due">{formatDateString(a.due_date)}</span> ({dateToDayOfWeek(a.due_date)})
                        {/if}
                    </div>
                </div>

                <div class="mobile-status">
                    <select value={a.status} on:input={(e) => statusAssignment(e, a.id)} style="background-color: {assignmentToColor(a)}">
                            <option value="Not Started">Not Started</option>
                            <option value="In Progress">In Progress</option>
                            <option value="Completed">Completed</option>
                    </select>
                </div>
            </div>
        {/each}
    {:else}
        <table>
            <tr>
                <th class="classWidth">Class</th>
                <th class="zeroWidth untightPadding">Type</th>
                <th>Name</th>
                <th class="zeroWidth untightPadding">Assigned</th>
                <th class="zeroWidth">Due</th>
                <th class="zeroWidth"></th>
                <th class="zeroWidth"></th>
            </tr>
            {#each shownAssignments as a, i (a.id)}
                {#key unique[a.id]}
                    <tr
                        in:fly|global={{ duration: 300, x: -200 }}
                        style="{shownAssignments[i + 1] && a.due_date != shownAssignments[i + 1].due_date ? 'border-bottom: 1px solid black;' : ''}"
                    >
                        <td class="breakWord padding">{classFromId(a.class_id).name}</td>
                        <td class="untightPadding">{a.type}</td>
                        <td>
                            <!-- svelte-ignore a11y-click-events-have-key-events -->
                            <!-- svelte-ignore a11y-no-static-element-interactions -->
                            <!-- svelte-ignore a11y-missing-attribute -->
                            <a class="breakWord pointer" on:click={() => assignmentDetailsButton(a.id)}>{a.name}</a>
                        </td>
                        <td class="untightPadding">{formatDateString(a.assigned_date)}</td>
                        {#if a.due_time != ""}
                            <td class="untightPadding">{formatDateString(a.due_date)} <br /> {dateToDayOfWeek(a.due_date)} <br /> {convertTimeTo12Hours(a.due_time)}</td>
                        {:else}
                            <td class="untightPadding">{formatDateString(a.due_date)} <br /> {dateToDayOfWeek(a.due_date)}</td>
                        {/if}

                        <td class="zeroWidth untightPadding">
                            <select value={a.status} on:input={(e) => statusAssignment(e, a.id)} style="background-color: {assignmentToColor(a)}">
                                <option value="Not Started">Not Started</option>
                                <option value="In Progress">In Progress</option>
                                <option value="Completed">Completed</option>
                            </select>
                        </td>

                        <td class="zeroWidth">
                            <button type="button" on:click={() => updateAssignmentButton(a.id)}>Edit</button>
                        </td>
                    </tr>
                {/key}
            {/each}
        </table>
    {/if}

    <br />
    <br />
    <br />
{/if}

</div>

<Modal bind:showModal={showICSModal}>
    <h2>Calender Integration (ICS)</h2>

    <TimezonePicker {timezone} on:update="{updateTimezone}" />
    
    {#if ics_link != ""}
        <p>Subscribe to your assignments in your calendar app!</p>

        <p class="calenderLabel">All Assignments:</p>
        <code class="breakWord">{`${api}/ics/0/${ics_link}.ics`}</code>

        <p class="calenderLabel">Not Started:</p>
        <code class="breakWord">{`${api}/ics/1/${ics_link}.ics`}</code>

        <p class="calenderLabel">In Progress:</p>
        <code class="breakWord">{`${api}/ics/2/${ics_link}.ics`}</code>

        <p class="calenderLabel">Completed:</p>
        <code class="breakWord">{`${api}/ics/3/${ics_link}.ics`}</code>

        <br />
    {:else}
        <p>No link generated yet.</p>
    {/if}

    <button id="generateICSLinkButton" type="button" disabled={!generateICSLinkButton} on:click={generateICSLink}>
        {#if ics_link == ""}
            Generate Link
        {:else}
            Regenerate Link
        {/if}
    </button>
</Modal>


<Modal bind:showModal={showCreateSemesterModal}>
    <h2>Create Semester</h2>
    <form id="createSemester">
        <label for="name">Name:</label>
        <input type="text" id="createSemesterModalName" name="name" required>
        <label for="sort_order">Sort Order:</label>
        <input type="number" id="createSemesterModalSortOrder" name="sort_order" value="{Math.max(...semesters.map(s => s.sort_order)) + 1}" required>
        
        <br />
        <button type="submit" disabled={!createSemesterModalButton}>Create</button>
    </form>
</Modal>

<Modal bind:showModal={showAllSemestersModal}>
    <h2>All Semesters</h2>
    <table>
        <tr>
            <th>Name</th>
            <th class="zeroWidth"></th>
            <th class="zeroWidth"></th>
        </tr>
        {#each semesters as s (s.id)}
            <tr
                in:fly|global={{ duration: 300, x: -200 }}
            >
                <td class="breakWord">{s.name}</td>
                <td class="zeroWidth">
                    <button type="button" on:click={() => updateSemesterButton(s.id)}>Edit</button>
                </td>
                <td class="zeroWidth">
                    <button type="button" on:click={() => deleteModalButtonSemester(s.id)} disabled={!deleteSemesterModalButton}>Delete</button>
                </td>
            </tr>
        {/each}
    </table>
    <button id="createSemesterButton" type="button" on:click={() => showCreateSemesterModal = true}>Create Semester</button>
</Modal>

<Modal bind:showModal={showUpdateSemesterModal}>
    <h2>Update Semester</h2>
    <form id="updateSemester">
        <label for="name">Name:</label>
        <input type="text" id="updateSemesterModalName" name="name" bind:value={updateSemesterModalName} required>
        <label for="sort_order">Sort Order:</label>
        <input type="number" id="updateSemesterModalSortOrder" name="sort_order" bind:value={updateSemesterModalSortOrder} required>
        <input type="hidden" name="id" bind:value={updateSemesterModalId}>
        
        <br />
        <button type="submit" disabled={!updateSemesterModalButton}>Update</button>
    </form>
</Modal>


<Modal bind:showModal={showCreateClassModal}>
    <h2>Create Class</h2>
    <form id="createClass">
        <label for="name">Name:</label>
        <input type="text" id="createClassModalName" name="name" required>
        <select id="createClassModalSemesterID" name="semester_id" required>
            {#each semesters as s (s.id)}
                <option value={s.id}>{s.name}</option>
            {/each}
        </select>
        
        <br />
        <button type="submit" disabled={!createClassModalButton}>Create</button>
    </form>
</Modal>

<Modal bind:showModal={showUpdateClassModal}>
    <h2>Update Class</h2>
    <form id="updateClass">
        <label for="name">Name:</label>
        <input type="text" id="updateClassModalName" name="name" bind:value={updateClassModalName} required>
        <select id="updateClassModalSemesterID" name="semester_id" bind:value={updateClassModalSemesterID} required>
            {#each semesters as s (s.id)}
                <option value={s.id}>{s.name}</option>
            {/each}
        </select>
        <input type="hidden" name="id" bind:value={updateClassModalId}>
        
        <br />
        <button type="submit" disabled={!updateClassModalButton}>Update</button>
    </form>
</Modal>

<Modal bind:showModal={showDeleteClassModal}>
    <h2>Delete Class</h2>
    <p>Are you sure you want to delete the class "{deleteClassModalName}"?</p>
    <p>Deleting the class will delete all assignments associated with it.</p>
    {#if deleteClassModalTimer > 0}
        <p>Wait {deleteClassModalTimer} seconds before deleting.</p>
    {:else}
        <button type="button" on:click={() => deleteClassButton(deleteClassModalId)} disabled={!deleteClassModalButton}>Yes</button>
    {/if}
</Modal>


<Modal bind:showModal={showCreateAssignmentModal}>
    <h2>Create Assignment</h2>
    <form id="createAssignment">
        <label for="name">Name:</label>
        <input type="text" id="createAssignmentModalName" name="name" required>
        <br />
        
        <label for="description">Description:</label>
        <textarea id="createAssignmentModalDescription" name="description"></textarea>
        
        <label for="assigned_date">Assigned Date:</label>
        <input type="date" id="createAssignmentModalAssignedDate" name="assigned_date" required>
        <br />

        <label for="due_date">Due Date:</label>
        <input type="date" id="createAssignmentModalDueDate" name="due_date" required>
        <br />

        <label for="due_time">Due Time:</label>
        <input type="time" id="createAssignmentModalDueTime" name="due_time">
        <br />
        
        <label for="status">Status:</label>
        <select id="createAssignmentModalStatus" name="status">
            <option value="Not Started" selected="selected">Not Started</option>
            <option value="In Progress">In Progress</option>
            <option value="Completed">Completed</option>
        </select>
        <br />

        <label for="type">Type:</label>
        <select id="createAssignmentModalType" name="type">
            <option value="Homework" selected="selected">Homework</option>
            <option value="Quiz">Quiz</option>
            <option value="Test">Test</option>
            <option value="Project">Project</option>
            <option value="Paper">Paper</option>
            <option value="Other">Other</option>
        </select>
        <br />

        <label for="class_id">Class:</label>
        <select id="createAssignmentModalClassId" name="class_id">
            {#each shownClasses as c (c.id)}
                <option value={c.id}>{c.name}</option>
            {/each}
        </select>

        <br />
        <button type="submit" disabled={!createAssignmentModalButton}>Create</button>
    </form>
</Modal>

<Modal bind:showModal={showUpdateAssignmentModal}>
    <h2>Update Assignment</h2>
    <form id="updateAssignment">
        <label for="name">Name:</label>
        <input type="text" id="updateAssignmentModalName" name="name" bind:value={updateAssignmentModalName} required>
        <br />
        
        <label for="description">Description:</label>
        <textarea id="updateAssignmentModalDescription" name="description" bind:value={updateAssignmentModalDescription}></textarea>
        
        <label for="assigned_date">Assigned Date:</label>
        <input type="date" id="updateAssignmentModalAssignedDate" name="assigned_date" bind:value={updateAssignmentModalAssignedDate} required>
        <br />

        <label for="due_date">Due Date:</label>
        <input type="date" id="updateAssignmentModalDueDate" name="due_date" bind:value={updateAssignmentModalDueDate} required>
        <br />

        <label for="due_time">Due Time:</label>
        <input type="time" id="updateAssignmentModalDueTime" name="due_time" bind:value={updateAssignmentModalDueTime}>
        <br />
        
        <label for="status">Status:</label>
        <select id="updateAssignmentModalStatus" name="status" bind:value={updateAssignmentModalStatus}>
            <option value="Not Started" selected="selected">Not Started</option>
            <option value="In Progress">In Progress</option>
            <option value="Completed">Completed</option>
        </select>
        <br />

        <label for="type">Type:</label>
        <select id="updateAssignmentModalType" name="type" bind:value={updateAssignmentModalType}>
            <option value="Homework" selected="selected">Homework</option>
            <option value="Quiz">Quiz</option>
            <option value="Test">Test</option>
            <option value="Project">Project</option>
            <option value="Paper">Paper</option>
            <option value="Other">Other</option>
        </select>
        <br />

        <label for="class_id">Class:</label>
        <select id="updateAssignmentModalClassId" name="class_id" bind:value={updateAssignmentModalClassId}>
            {#each shownClasses as c (c.id)}
                <option value={c.id}>{c.name}</option>
            {/each}
        </select>

        <input type="hidden" name="id" value={updateAssignmentModalId}>

        <br />
        <button type="submit" disabled={!updateAssignmentModalButton}>Update</button>
        <button type="button" on:click={() => deleteAssignmentButton(updateAssignmentModalId)} disabled={!updateAssignmentModalDeleteButton}>Delete</button>
    </form>
</Modal>

<Modal bind:showModal={showAssignmentDetailsModal}>
    <h2>Assignment Details</h2>
    <p>Name: {assignmentDetailsModalName}</p>
    <p class="description">Description: <br/>{assignmentDetailsModalDescription}</p>
    <p>Assigned Date: {formatDateString(assignmentDetailsModalAssignedDate)}</p>

    {#if assignmentDetailsModalDueTime != ""}
        <p>Due Date: {formatDateString(assignmentDetailsModalDueDate)} - {convertTimeTo12Hours(assignmentDetailsModalDueTime)}</p>
    {:else}
        <p>Due Date: {formatDateString(assignmentDetailsModalDueDate)}</p>
    {/if}
    <p>Status: {assignmentDetailsModalStatus}</p>
    <p>Type: {assignmentDetailsModalType}</p>
    <p>Class: {classFromId(assignmentDetailsModalClassId)?.name}</p>
    <button type="button" on:click={() => updateAssignmentButton(assignmentDetailsModalId)}>Edit</button>	
</Modal>


<Modal bind:showModal={showDateFilterModal}>
    <h2>Date Filter</h2>

    <label for="all">
        <input type="radio" id="all" value="all" bind:group={dateFilter}>
        All
    </label>

    <label for="week">
        <input type="radio" id="week" value="week" bind:group={dateFilter}>
        This Week
    </label>

    <label for="future">
        <input type="radio" id="future" value="future" bind:group={dateFilter}>
        Future
    </label>

    <label for="range">
        <input type="radio" id="range" value="range" bind:group={dateFilter}>
        Custom Range
    </label>

    {#if dateFilter == "range"}
        <label for="dateStart">Start Date:</label>
        <input type="date" id="dateStart" bind:value={dateStart}>
        <label for="dateEnd">End Date:</label>
        <input type="date" id="dateEnd" bind:value={dateEnd}>
    {/if}

    {#if dateFilter == "range" || dateFilter == "week"}
        <label for="rangeIncludesAssigned">
            <input type="checkbox" id="rangeIncludesAssigned" bind:checked={rangeIncludesAssigned}>
            Include Assigned Dates
        </label>
    {/if}
</Modal>

<Modal bind:showModal={showClassFilterModal}>
    <h2>Class Filter</h2>
    {#each shownClasses as c (c.id)}
        <label for={c.id}>
            <input type="checkbox" id={c.id} value={c.id} bind:group={classFilter}>
            {c.name}
        </label>
    {/each}

    <button type="button" on:click={() => classFilter = classes.map((c) => c.id)}>Select All</button>
    <button type="button" on:click={() => classFilter = []}>Select None</button>
</Modal>

<Modal bind:showModal={showStatusFilter}>
    <h2>Status Filter</h2>
    {#each ["Not Started", "In Progress", "Completed"] as s}
        <label for={s}>
            <input type="checkbox" id={s} value={s} bind:group={statusFilter}>
            {s}
        </label>
    {/each}
</Modal>

<Modal bind:showModal={showTypeFilter}>
    <h2>Type Filter</h2>
    {#each ["Homework", "Quiz", "Test", "Project", "Paper", "Other"] as t}
        <label for={t}>
            <input type="checkbox" id={t} value={t} bind:group={typeFilter}>
            {t}
        </label>
    {/each}

    <button type="button" on:click={() => typeFilter = ["Homework", "Quiz", "Test", "Project", "Paper", "Other"]}>Select All</button>
    <button type="button" on:click={() => typeFilter = []}>Select None</button>
</Modal>